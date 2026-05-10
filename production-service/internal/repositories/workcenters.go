package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListWorkcenters список.
func (s *Store) ListWorkcenters(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Workcenter, error) {
	q := `SELECT id, tenant_code, code, name, active, capacity_minutes_per_shift, created_at
		FROM workcenters WHERE tenant_code=$1 ORDER BY code`
	rows, err := s.db(tx).Query(ctx, q, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Workcenter
	for rows.Next() {
		var w models.Workcenter
		if err := rows.Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.Active, &w.CapacityMinutesPerShift, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GetWorkcenter возвращает по id.
func (s *Store) GetWorkcenter(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Workcenter, error) {
	q := `SELECT id, tenant_code, code, name, active, capacity_minutes_per_shift, created_at
		FROM workcenters WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var w models.Workcenter
	if err := row.Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.Active, &w.CapacityMinutesPerShift, &w.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

// CreateWorkcenter создаёт.
func (s *Store) CreateWorkcenter(ctx context.Context, tx pgx.Tx, w *models.Workcenter) error {
	q := `INSERT INTO workcenters(id, tenant_code, code, name, active, capacity_minutes_per_shift)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, w.ID, w.TenantCode, w.Code, w.Name, w.Active, w.CapacityMinutesPerShift)
	return err
}

// UpdateWorkcenter обновляет.
func (s *Store) UpdateWorkcenter(ctx context.Context, tx pgx.Tx, w *models.Workcenter) error {
	q := `UPDATE workcenters SET code=$2, name=$3, active=$4, capacity_minutes_per_shift=$5 WHERE tenant_code=$6 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, w.ID, w.Code, w.Name, w.Active, w.CapacityMinutesPerShift, w.TenantCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteWorkcenter удаляет.
func (s *Store) DeleteWorkcenter(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	q := `DELETE FROM workcenters WHERE tenant_code=$1 AND id=$2`
	tag, err := s.db(tx).Exec(ctx, q, tenant, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
