package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListScrapReasons список.
func (s *Store) ListScrapReasons(ctx context.Context, tx pgx.Tx, tenant string) ([]models.ScrapReason, error) {
	q := `SELECT id, tenant_code, code, name FROM scrap_reasons WHERE tenant_code=$1 ORDER BY code`
	rows, err := s.db(tx).Query(ctx, q, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ScrapReason
	for rows.Next() {
		var r models.ScrapReason
		if err := rows.Scan(&r.ID, &r.TenantCode, &r.Code, &r.Name); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CreateScrapReason создаёт.
func (s *Store) CreateScrapReason(ctx context.Context, tx pgx.Tx, r *models.ScrapReason) error {
	q := `INSERT INTO scrap_reasons(id, tenant_code, code, name) VALUES ($1,$2,$3,$4)`
	_, err := s.db(tx).Exec(ctx, q, r.ID, r.TenantCode, r.Code, r.Name)
	return err
}

// GetScrapReason по коду.
func (s *Store) GetScrapReason(ctx context.Context, tx pgx.Tx, tenant, code string) (*models.ScrapReason, error) {
	q := `SELECT id, tenant_code, code, name FROM scrap_reasons WHERE tenant_code=$1 AND code=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, code)
	var r models.ScrapReason
	if err := row.Scan(&r.ID, &r.TenantCode, &r.Code, &r.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}
