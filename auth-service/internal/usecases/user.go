package usecases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/models"
	"github.com/industrial-sed/auth-service/internal/ports"
)

// UserUC управление пользователями в рамках тенанта.
type UserUC struct {
	kc     ports.KeycloakClient
	repo   TenantRepository
	users  UserCacheRepository
	adm    *adminTokenSource
	realm  string
	notif  Notifier
}

// NewUserUC конструктор.
func NewUserUC(kc ports.KeycloakClient, tenants TenantRepository, users UserCacheRepository, n Notifier) *UserUC {
	return &UserUC{kc: kc, repo: tenants, users: users, adm: newAdminTokenSource(kc), notif: n}
}

// CreateUser создаёт пользователя username@tenant с временным паролем.
func (u *UserUC) CreateUser(ctx context.Context, actor *models.Claims, username, email, role string) (keycloakID, tempPassword string, err error) {
	if actor == nil || !actor.HasRole(keycloak.RoleEntAdmin) {
		return "", "", ErrForbidden
	}
	tenant := strings.TrimSpace(strings.ToLower(actor.TenantID))
	if tenant == "" {
		return "", "", fmt.Errorf("%w: tenant_id missing in token", ErrValidation)
	}
	if !allowedRole(role) {
		return "", "", fmt.Errorf("%w: invalid role", ErrValidation)
	}
	token, err := u.adm.Token(ctx)
	if err != nil {
		return "", "", err
	}
	tent, err := u.repo.GetByCode(ctx, tenant)
	if err != nil || tent == nil {
		return "", "", ErrNotFound
	}
	login := username + "@" + tenant
	if n, err := u.kc.CountUsersByUsername(ctx, token, login); err != nil {
		return "", "", err
	} else if n > 0 {
		return "", "", ErrConflict
	}
	tempPassword = randomPassword(16)
	enabled := true
	attrs := map[string][]string{"tenant_id": {tenant}}
	uRep := gocloak.User{
		Username:      gocloak.StringP(login),
		FirstName:     gocloak.StringP(username),
		LastName:      gocloak.StringP(tenant),
		Email:         gocloak.StringP(email),
		Enabled:       &enabled,
		EmailVerified: gocloak.BoolP(true),
		Attributes:    &attrs,
	}
	uid, err := u.kc.CreateUser(ctx, token, uRep)
	if err != nil {
		return "", "", fmt.Errorf("create user: %w", err)
	}
	// temporary=false: иначе Keycloak вешает required action «Update Password» и
	// password grant (в т.ч. /internal/test/login) отвечает invalid_grant «Account is not fully set up».
	if err := u.kc.SetUserPassword(ctx, token, uid, tempPassword, false); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", "", err
	}
	if err := u.kc.AddUserToGroup(ctx, token, uid, tent.KeycloakGroupID); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", "", err
	}
	rr, err := u.kc.RealmRole(ctx, token, role)
	if err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", "", err
	}
	if err := u.kc.AddRealmRoleToUser(ctx, token, uid, []gocloak.Role{*rr}); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", "", err
	}
	cache := &models.UserCache{
		KeycloakID: uid,
		TenantCode: tenant,
		Username:   login,
		Email:      email,
		Roles:      []string{role},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := u.users.Upsert(ctx, cache); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		return "", "", err
	}
	if err := u.notif.NotifyUserCreated(ctx, UserCreatedPayload{
		TenantCode:        tenant,
		Username:          login,
		Email:             email,
		TemporaryPassword: tempPassword,
		KeycloakUserID:    uid,
	}); err != nil {
		_ = u.kc.DeleteUser(ctx, token, uid)
		_ = u.users.Delete(ctx, uid)
		return "", "", err
	}
	return uid, tempPassword, nil
}

func allowedRole(r string) bool {
	switch r {
	case keycloak.RoleEntAdmin, keycloak.RoleApprover, keycloak.RoleEngineer, keycloak.RoleViewer,
		keycloak.RoleWarehouseAdmin, keycloak.RoleStorekeeper, keycloak.RoleWarehouseViewer,
		keycloak.RoleSedAdmin, keycloak.RoleSedAuthor, keycloak.RoleSedApprover, keycloak.RoleSedViewer,
		keycloak.RoleProdAdmin, keycloak.RoleProdTechnologist, keycloak.RoleProdPlanner, keycloak.RoleProdMaster,
		keycloak.RoleProdWorker, keycloak.RoleProdQC, keycloak.RoleProdViewer,
		keycloak.RoleProcAdmin, keycloak.RoleProcBuyer, keycloak.RoleProcApprover, keycloak.RoleProcViewer:
		return true
	default:
		return false
	}
}

// ListUsers пользователи текущего тенанта (кэш + сверка с KC).
func (u *UserUC) ListUsers(ctx context.Context, actor *models.Claims) ([]models.UserCache, error) {
	if actor == nil || actor.TenantID == "" {
		return nil, ErrForbidden
	}
	if !actor.HasRole(keycloak.RoleEntAdmin) {
		return nil, ErrForbidden
	}
	return u.users.ListByTenant(ctx, actor.TenantID)
}

// UpdateUserRoles заменяет realm-роли пользователя (только в своём тенанте).
func (u *UserUC) UpdateUserRoles(ctx context.Context, actor *models.Claims, userID string, roles []string) error {
	if actor == nil || !actor.HasRole(keycloak.RoleEntAdmin) {
		return ErrForbidden
	}
	for _, r := range roles {
		if r == keycloak.RoleSuperAdmin {
			return ErrForbidden
		}
		if !allowedRole(r) && r != "" {
			return fmt.Errorf("%w: role %q", ErrValidation, r)
		}
	}
	token, err := u.adm.Token(ctx)
	if err != nil {
		return err
	}
	if err := u.ensureUserInTenant(ctx, token, actor.TenantID, userID); err != nil {
		return err
	}
	var kcRoles []gocloak.Role
	for _, r := range roles {
		if r == "" {
			continue
		}
		rr, err := u.kc.RealmRole(ctx, token, r)
		if err != nil {
			return err
		}
		kcRoles = append(kcRoles, *rr)
	}
	if err := u.kc.SetRealmRolesForUser(ctx, token, userID, kcRoles); err != nil {
		return err
	}
	cached, err := u.users.GetByKeycloakID(ctx, userID)
	if err != nil || cached == nil {
		return nil
	}
	cached.Roles = roles
	return u.users.Upsert(ctx, cached)
}

// DeleteUser удаляет пользователя из Keycloak и кэша.
func (u *UserUC) DeleteUser(ctx context.Context, actor *models.Claims, userID string) error {
	if actor == nil || !actor.HasRole(keycloak.RoleEntAdmin) {
		return ErrForbidden
	}
	token, err := u.adm.Token(ctx)
	if err != nil {
		return err
	}
	if err := u.ensureUserInTenant(ctx, token, actor.TenantID, userID); err != nil {
		return err
	}
	if err := u.kc.DeleteUser(ctx, token, userID); err != nil {
		return err
	}
	return u.users.Delete(ctx, userID)
}

func (u *UserUC) ensureUserInTenant(ctx context.Context, token, tenantCode, userID string) error {
	tent, err := u.repo.GetByCode(ctx, tenantCode)
	if err != nil || tent == nil {
		return ErrNotFound
	}
	members, err := u.kc.GetUsersInGroup(ctx, token, tent.KeycloakGroupID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.ID != nil && *m.ID == userID {
			return nil
		}
	}
	return ErrNotFound
}

func randomPassword(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	b := make([]byte, n)
	max := big.NewInt(int64(len(chars)))
	for i := range b {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			b[i] = chars[i%len(chars)]
			continue
		}
		b[i] = chars[v.Int64()]
	}
	return string(b)
}
