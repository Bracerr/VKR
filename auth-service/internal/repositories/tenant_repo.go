package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/auth-service/internal/models"
)

// TenantRepo реализует usecases.TenantRepository.
type TenantRepo struct {
	pool *pgxpool.Pool
}

// NewTenantRepo конструктор.
func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

// Create вставляет тенант.
func (r *TenantRepo) Create(ctx context.Context, t *models.Tenant) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tenants (code, name, keycloak_group_id, created_at)
		VALUES ($1, $2, $3, $4)
	`, t.Code, t.Name, t.KeycloakGroupID, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

// List возвращает все тенанты.
func (r *TenantRepo) List(ctx context.Context) ([]models.Tenant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, name, keycloak_group_id, created_at FROM tenants ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()
	var out []models.Tenant
	for rows.Next() {
		var t models.Tenant
		if err := rows.Scan(&t.Code, &t.Name, &t.KeycloakGroupID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetByCode по коду.
func (r *TenantRepo) GetByCode(ctx context.Context, code string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT code, name, keycloak_group_id, created_at FROM tenants WHERE code = $1
	`, code).Scan(&t.Code, &t.Name, &t.KeycloakGroupID, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return &t, nil
}

// Delete удаляет тенант (каскадно user_cache).
func (r *TenantRepo) Delete(ctx context.Context, code string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM tenants WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListByCodePrefix — для очистки тестовых данных (prefix test_).
func (r *TenantRepo) ListByCodePrefix(ctx context.Context, prefix string) ([]models.Tenant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, name, keycloak_group_id, created_at FROM tenants WHERE code LIKE $1 ORDER BY code
	`, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("query tenants by prefix: %w", err)
	}
	defer rows.Close()
	var out []models.Tenant
	for rows.Next() {
		var t models.Tenant
		if err := rows.Scan(&t.Code, &t.Name, &t.KeycloakGroupID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
