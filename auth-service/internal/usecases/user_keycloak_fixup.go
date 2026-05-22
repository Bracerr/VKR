package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v13"

	"github.com/industrial-sed/auth-service/internal/ports"
)

// finalizeKeycloakUser снимает required actions и заполняет профиль — иначе Keycloak 24+ даёт
// invalid_grant «Account is not fully set up» при password grant и браузерном входе.
func finalizeKeycloakUser(ctx context.Context, kc ports.KeycloakClient, token, userID, login, tenantCode, password string) error {
	full, err := kc.GetUserByID(ctx, token, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	short := login
	if i := strings.Index(login, "@"); i > 0 {
		short = login[:i]
	}
	if full.FirstName == nil || strings.TrimSpace(*full.FirstName) == "" {
		full.FirstName = gocloak.StringP(short)
	}
	if full.LastName == nil || strings.TrimSpace(*full.LastName) == "" {
		full.LastName = gocloak.StringP(tenantCode)
	}
	enabled := true
	full.Enabled = &enabled
	full.EmailVerified = gocloak.BoolP(true)
	full.RequiredActions = &[]string{}
	if err := kc.UpdateUser(ctx, token, *full); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	return kc.SetUserPassword(ctx, token, userID, password, false)
}

// RepairTenantPasswords синхронизирует пароли и профиль всех пользователей тенанта в Keycloak.
func (u *UserUC) RepairTenantPasswords(ctx context.Context, tenantCode, password string) (int, error) {
	tenantCode = strings.ToLower(strings.TrimSpace(tenantCode))
	password = strings.TrimSpace(password)
	if tenantCode == "" || password == "" {
		return 0, fmt.Errorf("%w: tenant and password required", ErrValidation)
	}
	if len(password) < 8 {
		return 0, fmt.Errorf("%w: password min 8 chars", ErrValidation)
	}
	token, err := u.adm.Token(ctx)
	if err != nil {
		return 0, err
	}
	list, err := u.users.ListByTenant(ctx, tenantCode)
	if err != nil {
		return 0, err
	}
	for _, uc := range list {
		if err := finalizeKeycloakUser(ctx, u.kc, token, uc.KeycloakID, uc.Username, tenantCode, password); err != nil {
			return 0, fmt.Errorf("%s: %w", uc.Username, err)
		}
	}
	return len(list), nil
}
