package usecases

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/auth-service/internal/ports"
)

// TestCleanup удаляет тенанты с заданным префиксом кода из БД и группы в Keycloak,
// затем удаляет пользователей realm, у которых username содержит prefix (например test_).
func TestCleanup(ctx context.Context, prefix string, kc ports.KeycloakClient, tenants TenantRepository, users UserCacheRepository) error {
	if prefix == "" {
		prefix = "test_"
	}
	adm := newAdminTokenSource(kc)
	token, err := adm.Token(ctx)
	if err != nil {
		return err
	}
	list, err := tenants.ListByCodePrefix(ctx, prefix)
	if err != nil {
		return err
	}
	for _, t := range list {
		_ = users.DeleteByTenant(ctx, t.Code)
		_ = kc.DeleteGroup(ctx, token, t.KeycloakGroupID)
		if err := tenants.Delete(ctx, t.Code); err != nil && err != pgx.ErrNoRows {
			return err
		}
	}
	// Повторно получаем admin token после долгих операций
	token, err = adm.Token(ctx)
	if err != nil {
		return err
	}
	uu, err := kc.FindUsers(ctx, token, prefix)
	if err != nil {
		return err
	}
	pfx := strings.ToLower(prefix)
	for _, u := range uu {
		if u.ID != nil && u.Username != nil && strings.HasPrefix(strings.ToLower(*u.Username), pfx) {
			_ = kc.DeleteUser(ctx, token, *u.ID)
		}
	}
	return nil
}
