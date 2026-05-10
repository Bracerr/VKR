package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v13"

	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/models"
)

// CreateEntAdminBySuper создаёт первого администратора предприятия (только сценарий bootstrap, вызывает handler для super_admin).
func (u *UserUC) CreateEntAdminBySuper(ctx context.Context, tenantCode, username, email, password string) (userID string, err error) {
	tenantCode = strings.ToLower(strings.TrimSpace(tenantCode))
	token, err := u.adm.Token(ctx)
	if err != nil {
		return "", err
	}
	tent, err := u.repo.GetByCode(ctx, tenantCode)
	if err != nil || tent == nil {
		return "", ErrNotFound
	}
	login := username + "@" + tenantCode
	if n, err := u.kc.CountUsersByUsername(ctx, token, login); err != nil {
		return "", err
	} else if n > 0 {
		return "", ErrConflict
	}
	enabled := true
	attrs := map[string][]string{"tenant_id": {tenantCode}}
	uRep := gocloak.User{
		Username:      gocloak.StringP(login),
		FirstName:     gocloak.StringP(username),
		LastName:      gocloak.StringP(tenantCode),
		Email:         gocloak.StringP(email),
		Enabled:       &enabled,
		EmailVerified: gocloak.BoolP(true),
		Attributes:    &attrs,
	}
	uid, err := u.kc.CreateUser(ctx, token, uRep)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	if err := u.kc.SetUserPassword(ctx, token, uid, password, false); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", err
	}
	if err := u.kc.AddUserToGroup(ctx, token, uid, tent.KeycloakGroupID); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", err
	}
	rr, err := u.kc.RealmRole(ctx, token, keycloak.RoleEntAdmin)
	if err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", err
	}
	if err := u.kc.AddRealmRoleToUser(ctx, token, uid, []gocloak.Role{*rr}); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", err
	}
	cache := &models.UserCache{
		KeycloakID: uid,
		TenantCode: tenantCode,
		Username:   login,
		Email:      email,
		Roles:      []string{keycloak.RoleEntAdmin},
	}
	if err := u.users.Upsert(ctx, cache); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", err
	}
	return uid, nil
}
