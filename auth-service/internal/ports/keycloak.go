// Package ports содержит интерфейсы внешних зависимостей (без циклов импорта).
package ports

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

// TokenPair — access + refresh + id (опционально).
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	IDToken          string
	ExpiresIn        int
	RefreshExpiresIn int
}

// KeycloakClient — админ-операции и OIDC (обёртка над gocloak + HTTP).
type KeycloakClient interface {
	LoginAdmin(ctx context.Context) (*gocloak.JWT, error)
	EnsureRealmAndRoles(ctx context.Context, token string) error
	EnsureOAuthClient(ctx context.Context, token string, redirectURI string) (string, error)
	CreateGroup(ctx context.Context, token, name string, attrs map[string][]string) (string, error)
	DeleteGroup(ctx context.Context, token, groupID string) error
	GetGroupByPath(ctx context.Context, token, path string) (*gocloak.Group, error)
	CreateUser(ctx context.Context, token string, u gocloak.User) (string, error)
	SetUserPassword(ctx context.Context, token, userID, password string, temporary bool) error
	AddUserToGroup(ctx context.Context, token, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, token, userID, groupID string) error
	DeleteUser(ctx context.Context, token, userID string) error
	GetUsersInGroup(ctx context.Context, token, groupID string) ([]*gocloak.User, error)
	GetUserByID(ctx context.Context, token, userID string) (*gocloak.User, error)
	RealmRole(ctx context.Context, token, roleName string) (*gocloak.Role, error)
	AddRealmRoleToUser(ctx context.Context, token, userID string, roles []gocloak.Role) error
	SetRealmRolesForUser(ctx context.Context, token, userID string, roles []gocloak.Role) error
	GetRealmRolesOfUser(ctx context.Context, token, userID string) ([]*gocloak.Role, error)
	CountUsersByUsername(ctx context.Context, token, username string) (int, error)
	FindUsers(ctx context.Context, token, search string) ([]*gocloak.User, error)
	ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	LogoutRefreshToken(ctx context.Context, refreshToken string) error
	EndSessionURL(idTokenHint, postLogoutRedirect string) string
	PasswordGrant(ctx context.Context, username, password string) (*TokenPair, error)
	EnsureUserAttributeMapper(ctx context.Context, token string) error
}
