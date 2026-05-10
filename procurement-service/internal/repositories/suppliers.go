package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/procurement-service/internal/models"
)

func (s *Store) ListSuppliers(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Supplier, error) {
	q := `SELECT id, tenant_code, code, name, inn, kpp, contacts, active, created_at, updated_at
		FROM suppliers WHERE tenant_code=$1 ORDER BY code`
	rows, err := s.db(tx).Query(ctx, q, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Supplier
	for rows.Next() {
		var sp models.Supplier
		if err := rows.Scan(&sp.ID, &sp.TenantCode, &sp.Code, &sp.Name, &sp.INN, &sp.KPP, &sp.Contacts, &sp.Active, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *Store) GetSupplier(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Supplier, error) {
	q := `SELECT id, tenant_code, code, name, inn, kpp, contacts, active, created_at, updated_at
		FROM suppliers WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var sp models.Supplier
	if err := row.Scan(&sp.ID, &sp.TenantCode, &sp.Code, &sp.Name, &sp.INN, &sp.KPP, &sp.Contacts, &sp.Active, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

func (s *Store) CreateSupplier(ctx context.Context, tx pgx.Tx, sp *models.Supplier) error {
	q := `INSERT INTO suppliers(id, tenant_code, code, name, inn, kpp, contacts, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.db(tx).Exec(ctx, q, sp.ID, sp.TenantCode, sp.Code, sp.Name, sp.INN, sp.KPP, sp.Contacts, sp.Active)
	return err
}

func (s *Store) UpdateSupplier(ctx context.Context, tx pgx.Tx, sp *models.Supplier) error {
	q := `UPDATE suppliers SET code=$2, name=$3, inn=$4, kpp=$5, contacts=$6, active=$7, updated_at=now()
		WHERE tenant_code=$8 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, sp.ID, sp.Code, sp.Name, sp.INN, sp.KPP, sp.Contacts, sp.Active, sp.TenantCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteSupplier(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	q := `DELETE FROM suppliers WHERE tenant_code=$1 AND id=$2`
	tag, err := s.db(tx).Exec(ctx, q, tenant, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

