package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/industrial-sed/auth-service/internal/ports"
)

// Client обёртка над gocloak и token endpoint.
// baseURL — URL Keycloak для серверных запросов (в Docker часто http://keycloak:8080).
// browserBaseURL — URL Keycloak в браузере (например http://localhost:8081); для end_session.
type Client struct {
	gc               *gocloak.GoCloak
	baseURL          string
	browserBaseURL   string
	realm            string
	clientID         string
	clientSecret     string
	adminRealm       string
	adminUser        string
	adminPassword    string
	httpClient       *http.Client
}

// NewClient создаёт клиент Keycloak.
// browserBaseURL может быть пустым — тогда совпадает с baseURL.
func NewClient(baseURL, browserBaseURL, realm, clientID, clientSecret, adminRealm, adminUser, adminPassword string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	browserBaseURL = strings.TrimRight(strings.TrimSpace(browserBaseURL), "/")
	if browserBaseURL == "" {
		browserBaseURL = baseURL
	}
	return &Client{
		gc:             gocloak.NewClient(baseURL),
		baseURL:        baseURL,
		browserBaseURL: browserBaseURL,
		realm:          realm,
		clientID:       clientID,
		clientSecret:   clientSecret,
		adminRealm:     adminRealm,
		adminUser:      adminUser,
		adminPassword:  adminPassword,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

var _ ports.KeycloakClient = (*Client)(nil)

// LoginAdmin получает токен администратора master.
func (c *Client) LoginAdmin(ctx context.Context) (*gocloak.JWT, error) {
	return c.gc.LoginAdmin(ctx, c.adminUser, c.adminPassword, c.adminRealm)
}

// EnsureRealmAndRoles создаёт realm и роли при отсутствии.
func (c *Client) EnsureRealmAndRoles(ctx context.Context, token string) error {
	_, err := c.gc.GetRealm(ctx, token, c.realm)
	if err == nil {
		return c.ensureRealmRoles(ctx, token)
	}
	enabled := true
	realm := gocloak.RealmRepresentation{
		Realm:   gocloak.StringP(c.realm),
		Enabled: &enabled,
	}
	if _, err := c.gc.CreateRealm(ctx, token, realm); err != nil {
		return fmt.Errorf("create realm: %w", err)
	}
	return c.ensureRealmRoles(ctx, token)
}

func (c *Client) ensureRealmRoles(ctx context.Context, token string) error {
	roles, err := c.gc.GetRealmRoles(ctx, token, c.realm, gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("list realm roles: %w", err)
	}
	existing := map[string]struct{}{}
	for _, r := range roles {
		if r.Name != nil {
			existing[*r.Name] = struct{}{}
		}
	}
	for _, name := range RealmRoles {
		if _, ok := existing[name]; ok {
			continue
		}
		role := gocloak.Role{Name: gocloak.StringP(name)}
		if _, err := c.gc.CreateRealmRole(ctx, token, c.realm, role); err != nil {
			return fmt.Errorf("create realm role %s: %w", name, err)
		}
	}
	return nil
}

// BuildPostLogoutRedirectURIs шаблоны для атрибута Keycloak post.logout.redirect.uris (разделитель ##).
func BuildPostLogoutRedirectURIs(frontendURL string) string {
	u, err := url.Parse(strings.TrimSpace(frontendURL))
	if err != nil || u.Hostname() == "" {
		return "http://localhost:3000/*"
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "http"
	}
	hostWithPort := u.Hostname()
	if u.Port() != "" {
		hostWithPort += ":" + u.Port()
	}
	base := scheme + "://" + hostWithPort
	patterns := []string{base + "/*"}
	if u.Hostname() == "localhost" && u.Port() != "" {
		patterns = append(patterns, fmt.Sprintf("http://127.0.0.1:%s/*", u.Port()))
	}
	if u.Hostname() == "127.0.0.1" && u.Port() != "" {
		patterns = append(patterns, fmt.Sprintf("http://localhost:%s/*", u.Port()))
	}
	return strings.Join(patterns, "##")
}

// mergeOAuthClientAttributes сохраняет прочие атрибуты клиента и задаёт OIDC/PKCE/logout.
func mergeOAuthClientAttributes(existing *map[string]string, postLogoutURIs string) *map[string]string {
	m := map[string]string{}
	if existing != nil {
		for k, v := range *existing {
			m[k] = v
		}
	}
	m["pkce.code.challenge.method"] = "S256"
	m["post.logout.redirect.uris"] = postLogoutURIs
	return &m
}

// EnsureOAuthClient создаёт/обновляет confidential OIDC-клиент.
func (c *Client) EnsureOAuthClient(ctx context.Context, token string, redirectURI string, postLogoutURIs string) (string, error) {
	clients, err := c.gc.GetClients(ctx, token, c.realm, gocloak.GetClientsParams{ClientID: gocloak.StringP(c.clientID)})
	if err != nil {
		return "", fmt.Errorf("get clients: %w", err)
	}
	pub := false
	direct := true
	std := true
	secret := c.clientSecret
	attrs := mergeOAuthClientAttributes(nil, postLogoutURIs)
	rep := gocloak.Client{
		ClientID:                  gocloak.StringP(c.clientID),
		Name:                      gocloak.StringP(c.clientID),
		Enabled:                   gocloak.BoolP(true),
		PublicClient:              &pub,
		DirectAccessGrantsEnabled: &direct,
		StandardFlowEnabled:         &std,
		RedirectURIs:              &[]string{redirectURI},
		WebOrigins:                &[]string{"+"},
		Secret:                    &secret,
		Attributes:                attrs,
	}
	if len(clients) == 0 {
		id, err := c.gc.CreateClient(ctx, token, c.realm, rep)
		if err != nil {
			return "", fmt.Errorf("create client: %w", err)
		}
		return id, nil
	}
	cid := *clients[0].ID
	full, err := c.gc.GetClient(ctx, token, c.realm, cid)
	if err != nil {
		return "", fmt.Errorf("get client: %w", err)
	}
	full.RedirectURIs = rep.RedirectURIs
	full.WebOrigins = rep.WebOrigins
	full.StandardFlowEnabled = rep.StandardFlowEnabled
	full.DirectAccessGrantsEnabled = rep.DirectAccessGrantsEnabled
	full.PublicClient = rep.PublicClient
	full.Secret = rep.Secret
	full.Attributes = mergeOAuthClientAttributes(full.Attributes, postLogoutURIs)
	if err := c.gc.UpdateClient(ctx, token, c.realm, *full); err != nil {
		return "", fmt.Errorf("update client: %w", err)
	}
	return cid, nil
}

// EnsureUserAttributeMapper добавляет mapper tenant_id из атрибута пользователя.
func (c *Client) EnsureUserAttributeMapper(ctx context.Context, token string) error {
	clients, err := c.gc.GetClients(ctx, token, c.realm, gocloak.GetClientsParams{ClientID: gocloak.StringP(c.clientID)})
	if err != nil || len(clients) == 0 {
		return fmt.Errorf("client not found for mapper: %w", err)
	}
	cid := *clients[0].ID
	full, err := c.gc.GetClient(ctx, token, c.realm, cid)
	if err != nil {
		return fmt.Errorf("get client for mappers: %w", err)
	}
	const mapperName = "tenant_id-user-attribute"
	if full.ProtocolMappers != nil {
		for _, m := range *full.ProtocolMappers {
			if m.Name != nil && *m.Name == mapperName {
				return nil
			}
		}
	}
	cfg := map[string]string{
		"user.attribute":       "tenant_id",
		"claim.name":           "tenant_id",
		"jsonType.label":       "String",
		"id.token.claim":       "true",
		"access.token.claim":   "true",
		"userinfo.token.claim": "true",
	}
	pm := gocloak.ProtocolMapperRepresentation{
		Name:           gocloak.StringP(mapperName),
		Protocol:       gocloak.StringP("openid-connect"),
		ProtocolMapper: gocloak.StringP("oidc-usermodel-attribute-mapper"),
		Config:         &cfg,
	}
	if _, err := c.gc.CreateClientProtocolMapper(ctx, token, c.realm, cid, pm); err != nil {
		return fmt.Errorf("create protocol mapper: %w", err)
	}
	return nil
}

// CreateGroup создаёт группу с атрибутами.
func (c *Client) CreateGroup(ctx context.Context, token, name string, attrs map[string][]string) (string, error) {
	g := gocloak.Group{Name: gocloak.StringP(name), Attributes: &attrs}
	id, err := c.gc.CreateGroup(ctx, token, c.realm, g)
	if err != nil {
		return "", err
	}
	return id, nil
}

// DeleteGroup удаляет группу по ID.
func (c *Client) DeleteGroup(ctx context.Context, token, groupID string) error {
	return c.gc.DeleteGroup(ctx, token, c.realm, groupID)
}

// GetGroupByPath ищет группу по пути (например /tenant_x).
func (c *Client) GetGroupByPath(ctx context.Context, token, path string) (*gocloak.Group, error) {
	groups, err := c.gc.GetGroups(ctx, token, c.realm, gocloak.GetGroupsParams{Search: gocloak.StringP(strings.Trim(path, "/"))})
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if g.Path != nil && *g.Path == path {
			return g, nil
		}
	}
	return nil, fmt.Errorf("group path %q not found", path)
}

// CreateUser создаёт пользователя, возвращает ID.
func (c *Client) CreateUser(ctx context.Context, token string, u gocloak.User) (string, error) {
	return c.gc.CreateUser(ctx, token, c.realm, u)
}

// SetUserPassword задаёт пароль.
func (c *Client) SetUserPassword(ctx context.Context, token, userID, password string, temporary bool) error {
	return c.gc.SetPassword(ctx, token, userID, c.realm, password, temporary)
}

// AddUserToGroup добавляет в группу.
func (c *Client) AddUserToGroup(ctx context.Context, token, userID, groupID string) error {
	return c.gc.AddUserToGroup(ctx, token, c.realm, userID, groupID)
}

// RemoveUserFromGroup убирает из группы.
func (c *Client) RemoveUserFromGroup(ctx context.Context, token, userID, groupID string) error {
	return c.gc.DeleteUserFromGroup(ctx, token, c.realm, userID, groupID)
}

// DeleteUser удаляет пользователя.
func (c *Client) DeleteUser(ctx context.Context, token, userID string) error {
	return c.gc.DeleteUser(ctx, token, c.realm, userID)
}

// GetUsersInGroup пользователи в группе.
func (c *Client) GetUsersInGroup(ctx context.Context, token, groupID string) ([]*gocloak.User, error) {
	return c.gc.GetGroupMembers(ctx, token, c.realm, groupID, gocloak.GetGroupsParams{})
}

// GetUserByID получает пользователя.
func (c *Client) GetUserByID(ctx context.Context, token, userID string) (*gocloak.User, error) {
	return c.gc.GetUserByID(ctx, token, c.realm, userID)
}

// UpdateUser обновляет представление пользователя (например, сбрасывает requiredActions).
func (c *Client) UpdateUser(ctx context.Context, token string, user gocloak.User) error {
	return c.gc.UpdateUser(ctx, token, c.realm, user)
}

// RealmRole возвращает роль по имени.
func (c *Client) RealmRole(ctx context.Context, token, roleName string) (*gocloak.Role, error) {
	return c.gc.GetRealmRole(ctx, token, c.realm, roleName)
}

// AddRealmRoleToUser добавляет роли realm.
func (c *Client) AddRealmRoleToUser(ctx context.Context, token, userID string, roles []gocloak.Role) error {
	return c.gc.AddRealmRoleToUser(ctx, token, c.realm, userID, roles)
}

// SetRealmRolesForUser заменяет роли (удалить все кроме базовых не делаем — только назначенные).
func (c *Client) SetRealmRolesForUser(ctx context.Context, token, userID string, roles []gocloak.Role) error {
	current, err := c.gc.GetRealmRolesByUserID(ctx, token, c.realm, userID)
	if err != nil {
		return err
	}
	if len(current) > 0 {
		var del []gocloak.Role
		for _, r := range current {
			if r.Name != nil {
				del = append(del, gocloak.Role{ID: r.ID, Name: r.Name})
			}
		}
		if err := c.gc.DeleteRealmRoleFromUser(ctx, token, c.realm, userID, del); err != nil {
			return err
		}
	}
	if len(roles) > 0 {
		return c.gc.AddRealmRoleToUser(ctx, token, c.realm, userID, roles)
	}
	return nil
}

// GetRealmRolesOfUser список ролей пользователя.
func (c *Client) GetRealmRolesOfUser(ctx context.Context, token, userID string) ([]*gocloak.Role, error) {
	return c.gc.GetRealmRolesByUserID(ctx, token, c.realm, userID)
}

// UsersByExactUsername пользователи с точным совпадением username (для bootstrap и проверок).
func (c *Client) UsersByExactUsername(ctx context.Context, token, username string) ([]*gocloak.User, error) {
	return c.gc.GetUsers(ctx, token, c.realm, gocloak.GetUsersParams{
		Username: gocloak.StringP(username),
		Exact:    gocloak.BoolP(true),
	})
}

// CountUsersByUsername проверяет уникальность username в realm.
func (c *Client) CountUsersByUsername(ctx context.Context, token, username string) (int, error) {
	users, err := c.gc.GetUsers(ctx, token, c.realm, gocloak.GetUsersParams{Username: gocloak.StringP(username), Exact: gocloak.BoolP(true)})
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

// FindUsers поиск пользователей (для test cleanup).
func (c *Client) FindUsers(ctx context.Context, token, search string) ([]*gocloak.User, error) {
	return c.gc.GetUsers(ctx, token, c.realm, gocloak.GetUsersParams{Search: gocloak.StringP(search)})
}

func tokenEndpoint(c *Client) string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.baseURL, c.realm)
}

func logoutURL(c *Client) string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", c.baseURL, c.realm)
}

// ExchangeCode обменивает code на токены (PKCE).
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*ports.TokenPair, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)
	return c.postToken(ctx, form)
}

// RefreshToken обновляет access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*ports.TokenPair, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("refresh_token", refreshToken)
	return c.postToken(ctx, form)
}

// LogoutRefreshToken отзывает refresh на стороне KC (если поддерживается).
func (c *Client) LogoutRefreshToken(ctx context.Context, refreshToken string) error {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("refresh_token", refreshToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURL(c), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout: %s: %s", resp.Status, string(b))
	}
	return nil
}

// PasswordGrant для тестового логина.
func (c *Client) PasswordGrant(ctx context.Context, username, password string) (*ports.TokenPair, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("username", username)
	form.Set("password", password)
	return c.postToken(ctx, form)
}

func (c *Client) postToken(ctx context.Context, form url.Values) (*ports.TokenPair, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint(c), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s", PublicTokenErrorMessage(resp.StatusCode, body))
	}
	var tr struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		IDToken          string `json:"id_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}
	return &ports.TokenPair{
		AccessToken:      tr.AccessToken,
		RefreshToken:     tr.RefreshToken,
		IDToken:          tr.IDToken,
		ExpiresIn:        tr.ExpiresIn,
		RefreshExpiresIn: tr.RefreshExpiresIn,
	}, nil
}

// EndSessionURL URL для завершения SSO-сессии в браузере.
func (c *Client) EndSessionURL(idTokenHint, postLogoutRedirect string) string {
	u, _ := url.Parse(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", c.browserBaseURL, c.realm))
	q := u.Query()
	if idTokenHint != "" {
		q.Set("id_token_hint", idTokenHint)
	}
	q.Set("post_logout_redirect_uri", postLogoutRedirect)
	q.Set("client_id", c.clientID)
	u.RawQuery = q.Encode()
	return u.String()
}
