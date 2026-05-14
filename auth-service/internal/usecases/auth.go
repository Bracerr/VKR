package usecases

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/industrial-sed/auth-service/internal/models"
	"github.com/industrial-sed/auth-service/internal/ports"
)

// AuthUC OIDC code flow + PKCE (BFF).
type AuthUC struct {
	kc               ports.KeycloakClient
	issuer           string
	clientID         string
	callbackRedirect string
	cookieSecret     []byte
	frontendURL      string
	users            UserCacheRepository
}

// NewAuthUC конструктор.
// keycloakPublicURL — базовый URL Keycloak в браузере (issuer и /authorize для SPA).
func NewAuthUC(kc ports.KeycloakClient, keycloakPublicURL, realm, clientID, apiPublicURL, frontendURL, cookieSecret string, users UserCacheRepository) *AuthUC {
	pub := strings.TrimRight(keycloakPublicURL, "/")
	issuer := fmt.Sprintf("%s/realms/%s", pub, realm)
	cb := strings.TrimRight(apiPublicURL, "/") + "/api/v1/auth/callback"
	return &AuthUC{
		kc:               kc,
		issuer:           issuer,
		clientID:         clientID,
		callbackRedirect: cb,
		cookieSecret:     []byte(cookieSecret),
		frontendURL:      strings.TrimRight(frontendURL, "/"),
		users:            users,
	}
}

// FrontendBase URL фронтенда.
func (a *AuthUC) FrontendBase() string {
	return a.frontendURL
}

// returnToPath нормализует return_to до пути на фронте (/...), в т.ч. из полного URL того же origin.
func (a *AuthUC) returnToPath(returnTo string) string {
	rt := strings.TrimSpace(returnTo)
	if rt == "" {
		return "/"
	}
	if strings.HasPrefix(rt, "/") {
		return rt
	}
	u, err := url.Parse(rt)
	if err != nil || !u.IsAbs() {
		return "/" + strings.TrimPrefix(rt, "/")
	}
	f, err := url.Parse(a.frontendURL)
	if err != nil {
		return "/"
	}
	if !strings.EqualFold(u.Scheme, f.Scheme) || u.Host != f.Host {
		return "/"
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		path += "#" + u.Fragment
	}
	return path
}

// PostLoginRedirectURL абсолютный URL фронта после успешного входа.
func (a *AuthUC) PostLoginRedirectURL(returnTo string) string {
	base := strings.TrimRight(a.frontendURL, "/")
	path := a.returnToPath(returnTo)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

// oauthStatePayload данные в подписанном query-параметре OAuth state (без cookie).
// Cookie oauth_pkce ломалась при несовпадении хоста входа и redirect_uri (localhost vs LAN, 127.0.0.1).
type oauthStatePayload struct {
	Verifier  string `json:"v"`
	ReturnTo  string `json:"r"`
	ExpiresAt int64  `json:"e"`
}

// BuildAuthorizeURL генерирует URL редиректа на Keycloak (PKCE + подписанный state).
func (a *AuthUC) BuildAuthorizeURL(returnTo string) (authorizeURL string, err error) {
	verifier, err := randomVerifier(64)
	if err != nil {
		return "", err
	}
	if returnTo == "" {
		returnTo = "/"
	}
	returnTo = a.returnToPath(returnTo)
	if returnTo == "" {
		returnTo = "/"
	}
	pl := oauthStatePayload{
		Verifier:  verifier,
		ReturnTo:  returnTo,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
	}
	signedState, err := a.signPayload(pl)
	if err != nil {
		return "", err
	}
	challenge := pkceChallengeS256(verifier)
	q := url.Values{}
	q.Set("client_id", a.clientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", a.callbackRedirect)
	q.Set("state", signedState)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	authURL := a.issuer + "/protocol/openid-connect/auth?" + q.Encode()
	return authURL, nil
}

// ExchangeCallback проверяет подписанный state, обменивает code на токены.
func (a *AuthUC) ExchangeCallback(ctx context.Context, code, signedState string) (*ports.TokenPair, string, error) {
	pl, err := a.verifyPayload(signedState)
	if err != nil {
		return nil, "", err
	}
	if time.Now().Unix() > pl.ExpiresAt {
		return nil, "", fmt.Errorf("%w: state expired", ErrUnauthorized)
	}
	tokens, err := a.kc.ExchangeCode(ctx, code, pl.Verifier, a.callbackRedirect)
	if err != nil {
		return nil, "", err
	}
	return tokens, pl.ReturnTo, nil
}

// Refresh обновляет access token.
func (a *AuthUC) Refresh(ctx context.Context, refreshToken string) (*ports.TokenPair, error) {
	return a.kc.RefreshToken(ctx, refreshToken)
}

// LogoutSession отзывает refresh и возвращает URL завершения SSO.
func (a *AuthUC) LogoutSession(ctx context.Context, refreshToken, idToken string) (endSessionURL string, err error) {
	if refreshToken != "" {
		_ = a.kc.LogoutRefreshToken(ctx, refreshToken)
	}
	return a.kc.EndSessionURL(idToken, a.frontendURL+"/"), nil
}

// SyncUserCache обновляет кэш по claims после успешного логина.
func (a *AuthUC) SyncUserCache(ctx context.Context, c *models.Claims) error {
	if c == nil || c.Sub == "" || c.TenantID == "" {
		return nil
	}
	u := &models.UserCache{
		KeycloakID: c.Sub,
		TenantCode: c.TenantID,
		Username:   c.Username,
		Email:      c.Email,
		Roles:      c.RealmRoles,
	}
	return a.users.Upsert(ctx, u)
}

func randomVerifier(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	var sb strings.Builder
	sb.Grow(n)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[v.Int64()])
	}
	return sb.String(), nil
}

func pkceChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func (a *AuthUC) signPayload(pl oauthStatePayload) (string, error) {
	raw, err := json.Marshal(pl)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, a.cookieSecret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (a *AuthUC) verifyPayload(signed string) (*oauthStatePayload, error) {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: bad state", ErrUnauthorized)
	}
	payload, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, a.cookieSecret)
	mac.Write([]byte(payload))
	wantSig := mac.Sum(nil)
	gotSig, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(gotSig, wantSig) {
		return nil, fmt.Errorf("%w: bad signature", ErrUnauthorized)
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}
	var pl oauthStatePayload
	if err := json.Unmarshal(raw, &pl); err != nil {
		return nil, err
	}
	return &pl, nil
}
