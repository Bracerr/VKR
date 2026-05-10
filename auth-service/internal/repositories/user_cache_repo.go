package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/auth-service/internal/models"
)

// UserCacheRepo реализует usecases.UserCacheRepository.
type UserCacheRepo struct {
	pool *pgxpool.Pool
}

// NewUserCacheRepo конструктор.
func NewUserCacheRepo(pool *pgxpool.Pool) *UserCacheRepo {
	return &UserCacheRepo{pool: pool}
}

// Upsert вставляет или обновляет запись кэша.
func (r *UserCacheRepo) Upsert(ctx context.Context, u *models.UserCache) error {
	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_cache (keycloak_id, tenant_code, username, email, roles, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
		ON CONFLICT (keycloak_id) DO UPDATE SET
			tenant_code = EXCLUDED.tenant_code,
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			roles = EXCLUDED.roles,
			updated_at = EXCLUDED.updated_at
	`, u.KeycloakID, u.TenantCode, u.Username, u.Email, rolesJSON, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert user_cache: %w", err)
	}
	return nil
}

// ListByTenant пользователи тенанта.
func (r *UserCacheRepo) ListByTenant(ctx context.Context, tenantCode string) ([]models.UserCache, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT keycloak_id, tenant_code, username, email, roles, created_at, updated_at
		FROM user_cache WHERE tenant_code = $1 ORDER BY username
	`, tenantCode)
	if err != nil {
		return nil, fmt.Errorf("list user_cache: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func scanUsers(rows pgx.Rows) ([]models.UserCache, error) {
	var out []models.UserCache
	for rows.Next() {
		var u models.UserCache
		var rolesRaw []byte
		if err := rows.Scan(&u.KeycloakID, &u.TenantCode, &u.Username, &u.Email, &rolesRaw, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rolesRaw, &u.Roles); err != nil {
			u.Roles = nil
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetByKeycloakID одна запись.
func (r *UserCacheRepo) GetByKeycloakID(ctx context.Context, keycloakID string) (*models.UserCache, error) {
	var u models.UserCache
	var rolesRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT keycloak_id, tenant_code, username, email, roles, created_at, updated_at
		FROM user_cache WHERE keycloak_id = $1
	`, keycloakID).Scan(&u.KeycloakID, &u.TenantCode, &u.Username, &u.Email, &rolesRaw, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user_cache: %w", err)
	}
	_ = json.Unmarshal(rolesRaw, &u.Roles)
	return &u, nil
}

// Delete по keycloak id.
func (r *UserCacheRepo) Delete(ctx context.Context, keycloakID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_cache WHERE keycloak_id = $1`, keycloakID)
	return err
}

// DeleteByTenant все пользователи тенанта в кэше.
func (r *UserCacheRepo) DeleteByTenant(ctx context.Context, tenantCode string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_cache WHERE tenant_code = $1`, tenantCode)
	return err
}

// ListIDsCreatedBefore возвращает keycloak_id пользователей, созданных до указанного момента.
func (r *UserCacheRepo) ListIDsCreatedBefore(ctx context.Context, before time.Time, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, `
		SELECT keycloak_id
		FROM user_cache
		WHERE created_at < $1
		ORDER BY created_at ASC
		LIMIT $2
	`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("list user_cache ids: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
