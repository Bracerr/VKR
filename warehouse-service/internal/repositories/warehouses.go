package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// CreateWarehouse создаёт склад.
func (s *Store) CreateWarehouse(ctx context.Context, tx pgx.Tx, w *models.Warehouse) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO warehouses (id, tenant_code, code, name) VALUES ($1,$2,$3,$4)
	`, w.ID, w.TenantCode, w.Code, w.Name)
	return err
}

// GetWarehouseByID склад.
func (s *Store) GetWarehouseByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Warehouse, error) {
	var w models.Warehouse
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, code, name, created_at FROM warehouses WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

// ListWarehouses список.
func (s *Store) ListWarehouses(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Warehouse, error) {
	rows, err := s.db(tx).Query(ctx, `SELECT id, tenant_code, code, name, created_at FROM warehouses WHERE tenant_code = $1 ORDER BY code`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		if err := rows.Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// UpdateWarehouse обновление склада.
func (s *Store) UpdateWarehouse(ctx context.Context, tx pgx.Tx, w *models.Warehouse) error {
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE warehouses SET name = $3 WHERE id = $1 AND tenant_code = $2
	`, w.ID, w.TenantCode, w.Name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteWarehouse удаление.
func (s *Store) DeleteWarehouse(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM warehouses WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreateBin создаёт ячейку.
func (s *Store) CreateBin(ctx context.Context, tx pgx.Tx, b *models.Bin) error {
	var cap *string
	if b.CapacityQty != nil {
		v := b.CapacityQty.StringFixed(3)
		cap = &v
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO bins (id, tenant_code, warehouse_id, code, name, bin_type, parent_bin_id, capacity_qty)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, b.ID, b.TenantCode, b.WarehouseID, b.Code, b.Name, b.BinType, b.ParentBinID, cap)
	return err
}

// GetBinByID ячейка.
func (s *Store) GetBinByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Bin, error) {
	var b models.Bin
	var cap *string
	var parent *uuid.UUID
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, warehouse_id, code, name, bin_type, parent_bin_id, capacity_qty::text, created_at
		FROM bins WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&b.ID, &b.TenantCode, &b.WarehouseID, &b.Code, &b.Name, &b.BinType, &parent, &cap, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	b.ParentBinID = parent
	if cap != nil {
		d, err := decimal.NewFromString(*cap)
		if err == nil {
			b.CapacityQty = &d
		}
	}
	return &b, err
}

// ListBinsByWarehouse ячейки склада.
func (s *Store) ListBinsByWarehouse(ctx context.Context, tx pgx.Tx, tenant string, whID uuid.UUID) ([]models.Bin, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, warehouse_id, code, name, bin_type, parent_bin_id, capacity_qty::text, created_at
		FROM bins WHERE tenant_code = $1 AND warehouse_id = $2 ORDER BY code
	`, tenant, whID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Bin
	for rows.Next() {
		var b models.Bin
		var cap *string
		var parent *uuid.UUID
		if err := rows.Scan(&b.ID, &b.TenantCode, &b.WarehouseID, &b.Code, &b.Name, &b.BinType, &parent, &cap, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.ParentBinID = parent
		if cap != nil {
			d, err := decimal.NewFromString(*cap)
			if err == nil {
				b.CapacityQty = &d
			}
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// UpdateBin обновление.
func (s *Store) UpdateBin(ctx context.Context, tx pgx.Tx, b *models.Bin) error {
	var cap *string
	if b.CapacityQty != nil {
		v := b.CapacityQty.StringFixed(3)
		cap = &v
	}
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE bins SET name = $3, bin_type = $4, parent_bin_id = $5, capacity_qty = $6
		WHERE id = $1 AND tenant_code = $2
	`, b.ID, b.TenantCode, b.Name, b.BinType, b.ParentBinID, cap)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteBin удаление.
func (s *Store) DeleteBin(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM bins WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SumBinQuantity сумма quantity по ячейке (для capacity).
func (s *Store) SumBinQuantity(ctx context.Context, tx pgx.Tx, binID uuid.UUID) (decimal.Decimal, error) {
	var t *string
	err := s.db(tx).QueryRow(ctx, `SELECT COALESCE(SUM(quantity),0)::text FROM stock_balances WHERE bin_id = $1`, binID).Scan(&t)
	if err != nil {
		return decimal.Zero, err
	}
	if t == nil {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(*t)
}
