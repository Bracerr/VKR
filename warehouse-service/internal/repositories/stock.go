package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

func scanMovements(rows pgx.Rows) ([]models.StockMovement, error) {
	defer rows.Close()
	var out []models.StockMovement
	for rows.Next() {
		var m models.StockMovement
		var qs, ucs, vs *string
		if err := rows.Scan(&m.ID, &m.TenantCode, &m.MovementType, &m.DocumentID, &m.WarehouseID, &m.BinID, &m.ProductID, &m.BatchID, &m.SerialID,
			&qs, &ucs, &vs, &m.Currency, &m.PostedAt, &m.PostedBy); err != nil {
			return nil, err
		}
		m.Qty, _ = decimal.NewFromString(*qs)
		m.UnitCost = decPtr(ucs)
		m.Value = decPtr(vs)
		out = append(out, m)
	}
	return out, rows.Err()
}

// InsertMovement вставка движения.
func (s *Store) InsertMovement(ctx context.Context, tx pgx.Tx, m *models.StockMovement) error {
	var uc, val *string
	if m.UnitCost != nil {
		t := m.UnitCost.StringFixed(4)
		uc = &t
	}
	if m.Value != nil {
		t := m.Value.StringFixed(4)
		val = &t
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO stock_movements (id, tenant_code, movement_type, document_id, warehouse_id, bin_id, product_id, batch_id, serial_id, qty, unit_cost, value, currency, posted_at, posted_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, m.ID, m.TenantCode, m.MovementType, m.DocumentID, m.WarehouseID, m.BinID, m.ProductID, m.BatchID, m.SerialID, m.Qty.StringFixed(3), uc, val, m.Currency, m.PostedAt, m.PostedBy)
	return err
}

// LockBalance FOR UPDATE строки остатка.
func (s *Store) LockBalance(ctx context.Context, tx pgx.Tx, wh, bin, product, batch uuid.UUID) (*models.StockBalance, error) {
	var b models.StockBalance
	var qs, rs, vs string
	err := s.db(tx).QueryRow(ctx, `
		SELECT warehouse_id, bin_id, product_id, batch_id, quantity::text, reserved_qty::text, value::text, updated_at
		FROM stock_balances WHERE warehouse_id = $1 AND bin_id = $2 AND product_id = $3 AND batch_id = $4
		FOR UPDATE
	`, wh, bin, product, batch).Scan(&b.WarehouseID, &b.BinID, &b.ProductID, &b.BatchID, &qs, &rs, &vs, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.Quantity, _ = decimal.NewFromString(qs)
	b.ReservedQty, _ = decimal.NewFromString(rs)
	b.Value, _ = decimal.NewFromString(vs)
	return &b, nil
}

// UpsertBalanceDelta изменяет остаток (delta qty, delta reserved, delta value).
func (s *Store) UpsertBalanceDelta(ctx context.Context, tx pgx.Tx, wh, bin, product, batch uuid.UUID, dQty, dRes, dVal decimal.Decimal) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO stock_balances (warehouse_id, bin_id, product_id, batch_id, quantity, reserved_qty, value, updated_at)
		VALUES ($1,$2,$3,$4, $5, $6, $7, now())
		ON CONFLICT (warehouse_id, bin_id, product_id, batch_id) DO UPDATE SET
			quantity = stock_balances.quantity + EXCLUDED.quantity,
			reserved_qty = stock_balances.reserved_qty + EXCLUDED.reserved_qty,
			value = stock_balances.value + EXCLUDED.value,
			updated_at = now()
	`, wh, bin, product, batch, dQty.StringFixed(3), dRes.StringFixed(3), dVal.StringFixed(4))
	return err
}

// UpdateBalanceDeltaExisting изменяет остаток только если строка уже существует.
// Нужен, т.к. CHECK(reserved_qty <= quantity) может валидироваться для вставляемой строки
// до срабатывания ON CONFLICT, и резерв с dQty=0, dRes>0 приводит к 23514.
func (s *Store) UpdateBalanceDeltaExisting(ctx context.Context, tx pgx.Tx, wh, bin, product, batch uuid.UUID, dQty, dRes, dVal decimal.Decimal) error {
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE stock_balances SET
			quantity = quantity + $5,
			reserved_qty = reserved_qty + $6,
			value = value + $7,
			updated_at = now()
		WHERE warehouse_id = $1 AND bin_id = $2 AND product_id = $3 AND batch_id = $4
	`, wh, bin, product, batch, dQty.StringFixed(3), dRes.StringFixed(3), dVal.StringFixed(4))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListBalances остатки с фильтрами.
func (s *Store) ListBalances(ctx context.Context, tx pgx.Tx, tenant string, whID, binID, productID, batchID *uuid.UUID, onlyPos bool, expBefore *time.Time) ([]models.StockBalance, error) {
	q := `
		SELECT sb.warehouse_id, sb.bin_id, sb.product_id, sb.batch_id, sb.quantity::text, sb.reserved_qty::text, sb.value::text, sb.updated_at, b.expires_at
		FROM stock_balances sb
		JOIN warehouses w ON w.id = sb.warehouse_id
		LEFT JOIN batches b ON b.id = sb.batch_id
		WHERE w.tenant_code = $1`
	args := []interface{}{tenant}
	n := 2
	if whID != nil {
		q += ` AND sb.warehouse_id = $` + strconv.Itoa(n)
		args = append(args, *whID)
		n++
	}
	if binID != nil {
		q += ` AND sb.bin_id = $` + strconv.Itoa(n)
		args = append(args, *binID)
		n++
	}
	if productID != nil {
		q += ` AND sb.product_id = $` + strconv.Itoa(n)
		args = append(args, *productID)
		n++
	}
	if batchID != nil {
		q += ` AND sb.batch_id = $` + strconv.Itoa(n)
		args = append(args, *batchID)
		n++
	}
	if onlyPos {
		q += ` AND sb.quantity > 0`
	}
	if expBefore != nil {
		q += ` AND b.expires_at IS NOT NULL AND b.expires_at <= $` + strconv.Itoa(n)
		args = append(args, *expBefore)
	}
	q += ` ORDER BY sb.warehouse_id, sb.bin_id, sb.product_id`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.StockBalance
	for rows.Next() {
		var b models.StockBalance
		var qs, rs, vs string
		if err := rows.Scan(&b.WarehouseID, &b.BinID, &b.ProductID, &b.BatchID, &qs, &rs, &vs, &b.UpdatedAt, &b.ExpiresAt); err != nil {
			return nil, err
		}
		b.Quantity, _ = decimal.NewFromString(qs)
		b.ReservedQty, _ = decimal.NewFromString(rs)
		b.Value, _ = decimal.NewFromString(vs)
		out = append(out, b)
	}
	return out, rows.Err()
}
