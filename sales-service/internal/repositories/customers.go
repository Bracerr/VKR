package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/sales-service/internal/models"
)

func (s *Store) ListCustomers(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Customer, error) {
	q := `SELECT id, tenant_code, code, name, contacts, active, created_at, updated_at
		FROM customers WHERE tenant_code=$1 ORDER BY code`
	rows, err := s.db(tx).Query(ctx, q, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.TenantCode, &c.Code, &c.Name, &c.Contacts, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCustomer(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Customer, error) {
	q := `SELECT id, tenant_code, code, name, contacts, active, created_at, updated_at
		FROM customers WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var c models.Customer
	if err := row.Scan(&c.ID, &c.TenantCode, &c.Code, &c.Name, &c.Contacts, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateCustomer(ctx context.Context, tx pgx.Tx, c *models.Customer) error {
	q := `INSERT INTO customers(id, tenant_code, code, name, contacts, active)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, c.ID, c.TenantCode, c.Code, c.Name, c.Contacts, c.Active)
	return err
}

func (s *Store) UpdateCustomer(ctx context.Context, tx pgx.Tx, c *models.Customer) error {
	q := `UPDATE customers SET code=$2, name=$3, contacts=$4, active=$5, updated_at=now()
		WHERE tenant_code=$6 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, c.ID, c.Code, c.Name, c.Contacts, c.Active, c.TenantCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteCustomer(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	q := `DELETE FROM customers WHERE tenant_code=$1 AND id=$2`
	tag, err := s.db(tx).Exec(ctx, q, tenant, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

