package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/auth-service/internal/models"
	"github.com/industrial-sed/auth-service/internal/ports"
)

// TenantUC сценарии предприятий (суперадмин).
type TenantUC struct {
	kc     ports.KeycloakClient
	repo   TenantRepository
	adm    *adminTokenSource
	realm  string
}

// NewTenantUC конструктор.
func NewTenantUC(kc ports.KeycloakClient, repo TenantRepository) *TenantUC {
	return &TenantUC{kc: kc, repo: repo, adm: newAdminTokenSource(kc)}
}

// CreateTenant создаёт группу tenant_{code} в Keycloak и запись в БД.
func (u *TenantUC) CreateTenant(ctx context.Context, code, name string) (*models.Tenant, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return nil, fmt.Errorf("%w: code required", ErrValidation)
	}
	token, err := u.adm.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin token: %w", err)
	}
	groupName := "tenant_" + code
	path := "/" + groupName
	if g, err := u.kc.GetGroupByPath(ctx, token, path); err == nil && g != nil && g.ID != nil {
		return nil, fmt.Errorf("%w: tenant exists", ErrConflict)
	}
	attrs := map[string][]string{"tenant_id": {code}}
	gid, err := u.kc.CreateGroup(ctx, token, groupName, attrs)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	t := &models.Tenant{
		Code:            code,
		Name:            name,
		KeycloakGroupID: gid,
		CreatedAt:       time.Now().UTC(),
	}
	if err := u.repo.Create(ctx, t); err != nil {
		_ = u.kc.DeleteGroup(ctx, token, gid)
		return nil, err
	}
	return t, nil
}

// ListTenants все предприятия из локального кэша.
func (u *TenantUC) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	return u.repo.List(ctx)
}

// DeleteTenant удаляет группу и запись.
func (u *TenantUC) DeleteTenant(ctx context.Context, code string) error {
	token, err := u.adm.Token(ctx)
	if err != nil {
		return err
	}
	t, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrNotFound
	}
	if err := u.kc.DeleteGroup(ctx, token, t.KeycloakGroupID); err != nil {
		return fmt.Errorf("delete kc group: %w", err)
	}
	if err := u.repo.Delete(ctx, code); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
