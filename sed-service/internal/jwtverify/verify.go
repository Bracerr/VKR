// Package jwtverify проверяет access JWT Keycloak по JWKS.
package jwtverify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/industrial-sed/sed-service/internal/models"
)

type keyfuncJWKS interface {
	Keyfunc(token *jwt.Token) (any, error)
}

type KCAccessClaims struct {
	jwt.RegisteredClaims
	RealmAccess *struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	TenantID          string `json:"tenant_id"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Azp               string `json:"azp"`
}

// Parser проверяет JWT.
type Parser struct {
	clientID string
	jwks     keyfuncJWKS
}

// NewParser создаёт парсер JWKS.
func NewParser(ctx context.Context, keycloakBase, realm, clientID string) (*Parser, error) {
	base := strings.TrimRight(keycloakBase, "/")
	issuer := fmt.Sprintf("%s/realms/%s", base, realm)
	jwksURL := issuer + "/protocol/openid-connect/certs"
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("jwks: %w", err)
	}
	return &Parser{clientID: clientID, jwks: jwks}, nil
}

// ParseAccessToken валидирует токен и возвращает Claims.
func (p *Parser) ParseAccessToken(_ context.Context, raw string) (*models.Claims, error) {
	claims := &KCAccessClaims{}
	tok, err := jwt.ParseWithClaims(raw, claims, p.jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil || !tok.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !strings.Contains(claims.Issuer, "/realms/") {
		return nil, fmt.Errorf("wrong issuer")
	}
	if !claims.verifyAudience(p.clientID) {
		return nil, fmt.Errorf("wrong audience/azp")
	}
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("expired")
	}
	var roles []string
	if claims.RealmAccess != nil {
		roles = append(roles, claims.RealmAccess.Roles...)
	}
	un := claims.PreferredUsername
	if un == "" {
		un = claims.Subject
	}
	return &models.Claims{
		Sub:        claims.Subject,
		Username:   un,
		Email:      claims.Email,
		TenantID:   claims.TenantID,
		RealmRoles: roles,
	}, nil
}

func (c *KCAccessClaims) verifyAudience(clientID string) bool {
	if c.Azp == clientID {
		return true
	}
	for _, a := range c.Audience {
		if a == clientID {
			return true
		}
	}
	return false
}
