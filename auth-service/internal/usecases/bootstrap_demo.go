package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/models"
)

// BootstrapDemoPowerUser создаёт или обновляет пользователя с полным набором ролей тенанта
// (все RealmRoles кроме super_admin). Тенант должен существовать в БД; иначе — без ошибки.
func (u *UserUC) BootstrapDemoPowerUser(ctx context.Context, tenantCode, username, password string) error {
	tenantCode = strings.ToLower(strings.TrimSpace(tenantCode))
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if tenantCode == "" || username == "" || password == "" {
		return fmt.Errorf("%w: tenant, username and password required", ErrValidation)
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password min 8 chars", ErrValidation)
	}
	tent, err := u.repo.GetByCode(ctx, tenantCode)
	if err != nil || tent == nil {
		return nil
	}
	login := username + "@" + tenantCode
	roles := keycloak.TenantPowerRoles()
	token, err := u.adm.Token(ctx)
	if err != nil {
		return err
	}
	var uid string
	n, err := u.kc.CountUsersByUsername(ctx, token, login)
	if err != nil {
		return err
	}
	enabled := true
	attrs := map[string][]string{"tenant_id": {tenantCode}}
	email := login + ".demo.local"
	if n > 0 {
		users, err := u.kc.FindUsers(ctx, token, login)
		if err != nil {
			return fmt.Errorf("find demo user: %w", err)
		}
		for _, usr := range users {
			if usr.Username != nil && *usr.Username == login && usr.ID != nil {
				uid = *usr.ID
				break
			}
		}
		if uid == "" {
			return fmt.Errorf("demo user exists but cannot load")
		}
	} else {
		uRep := gocloak.User{
			Username:      gocloak.StringP(login),
			FirstName:     gocloak.StringP(username),
			LastName:      gocloak.StringP(tenantCode),
			Email:         gocloak.StringP(email),
			Enabled:       &enabled,
			EmailVerified: gocloak.BoolP(true),
			Attributes:    &attrs,
		}
		uid, err = u.kc.CreateUser(ctx, token, uRep)
		if err != nil {
			return fmt.Errorf("create demo user: %w", err)
		}
		if err := u.kc.AddUserToGroup(ctx, token, uid, tent.KeycloakGroupID); err != nil {
			_ = u.kc.DeleteUser(ctx, token, uid)
			return err
		}
	}
	if err := finalizeKeycloakUser(ctx, u.kc, token, uid, login, tenantCode, password); err != nil {
		return err
	}
	kcRoles := make([]gocloak.Role, 0, len(roles))
	for _, rn := range roles {
		rr, err := u.kc.RealmRole(ctx, token, rn)
		if err != nil {
			return fmt.Errorf("demo user role %s: %w", rn, err)
		}
		kcRoles = append(kcRoles, *rr)
	}
	if err := u.kc.SetRealmRolesForUser(ctx, token, uid, kcRoles); err != nil {
		return fmt.Errorf("demo user assign roles: %w", err)
	}
	cache := &models.UserCache{
		KeycloakID: uid,
		TenantCode: tenantCode,
		Username:   login,
		Email:      email,
		Roles:      roles,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if existing, _ := u.users.GetByKeycloakID(ctx, uid); existing != nil {
		cache.CreatedAt = existing.CreatedAt
	}
	return u.users.Upsert(ctx, cache)
}
