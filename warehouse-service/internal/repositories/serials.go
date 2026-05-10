package repositories

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// CreateSerial создаёт серийник.
func (s *Store) CreateSerial(ctx context.Context, tx pgx.Tx, sn *models.SerialNumber) error {
	var uc *string
	if sn.UnitCost != nil {
		v := sn.UnitCost.StringFixed(4)
		uc = &v
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO serial_numbers (id, tenant_code, product_id, batch_id, serial_no, status, warehouse_id, bin_id, unit_cost)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, sn.ID, sn.TenantCode, sn.ProductID, sn.BatchID, sn.SerialNo, sn.Status, sn.WarehouseID, sn.BinID, uc)
	return err
}

// GetSerialByProductNo серийник.
func (s *Store) GetSerialByProductNo(ctx context.Context, tx pgx.Tx, productID uuid.UUID, serialNo string) (*models.SerialNumber, error) {
	var sn models.SerialNumber
	var bid *uuid.UUID
	var wid *uuid.UUID
	var bid2 *uuid.UUID
	var uc *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, product_id, batch_id, serial_no, status, warehouse_id, bin_id, last_movement_id, unit_cost::text, created_at, updated_at
		FROM serial_numbers WHERE product_id = $1 AND serial_no = $2
	`, productID, serialNo).Scan(&sn.ID, &sn.TenantCode, &sn.ProductID, &bid, &sn.SerialNo, &sn.Status, &wid, &bid2, &sn.LastMovementID, &uc, &sn.CreatedAt, &sn.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	sn.BatchID = bid
	sn.WarehouseID = wid
	sn.BinID = bid2
	sn.UnitCost = decPtr(uc)
	return &sn, err
}

// GetSerialByID серийник по id.
func (s *Store) GetSerialByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.SerialNumber, error) {
	var sn models.SerialNumber
	var bid *uuid.UUID
	var wid *uuid.UUID
	var bid2 *uuid.UUID
	var uc *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, product_id, batch_id, serial_no, status, warehouse_id, bin_id, last_movement_id, unit_cost::text, created_at, updated_at
		FROM serial_numbers WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&sn.ID, &sn.TenantCode, &sn.ProductID, &bid, &sn.SerialNo, &sn.Status, &wid, &bid2, &sn.LastMovementID, &uc, &sn.CreatedAt, &sn.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	sn.BatchID = bid
	sn.WarehouseID = wid
	sn.BinID = bid2
	sn.UnitCost = decPtr(uc)
	return &sn, err
}

// UpdateSerialLocationAndStatus обновляет склад/ячейку/статус.
func (s *Store) UpdateSerialLocationAndStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, wh, bin *uuid.UUID, status string, movID *uuid.UUID) error {
	_, err := s.db(tx).Exec(ctx, `
		UPDATE serial_numbers SET warehouse_id = $2, bin_id = $3, status = $4, last_movement_id = $5, updated_at = now()
		WHERE id = $1
	`, id, wh, bin, status, movID)
	return err
}

// ListSerials фильтры.
func (s *Store) ListSerials(ctx context.Context, tx pgx.Tx, tenant string, productID *uuid.UUID, status *string, whID *uuid.UUID, binID *uuid.UUID, serialNo *string) ([]models.SerialNumber, error) {
	q := `SELECT id, tenant_code, product_id, batch_id, serial_no, status, warehouse_id, bin_id, last_movement_id, unit_cost::text, created_at, updated_at
		FROM serial_numbers WHERE tenant_code = $1`
	args := []interface{}{tenant}
	n := 2
	if productID != nil {
		q += ` AND product_id = $` + strconv.Itoa(n)
		args = append(args, *productID)
		n++
	}
	if status != nil {
		q += ` AND status = $` + strconv.Itoa(n)
		args = append(args, *status)
		n++
	}
	if whID != nil {
		q += ` AND warehouse_id = $` + strconv.Itoa(n)
		args = append(args, *whID)
		n++
	}
	if binID != nil {
		q += ` AND bin_id = $` + strconv.Itoa(n)
		args = append(args, *binID)
		n++
	}
	if serialNo != nil {
		q += ` AND serial_no ILIKE $` + strconv.Itoa(n)
		args = append(args, "%"+*serialNo+"%")
		n++
	}
	q += ` ORDER BY serial_no LIMIT 500`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SerialNumber
	for rows.Next() {
		var sn models.SerialNumber
		var bid *uuid.UUID
		var wid *uuid.UUID
		var bid2 *uuid.UUID
		var uc *string
		if err := rows.Scan(&sn.ID, &sn.TenantCode, &sn.ProductID, &bid, &sn.SerialNo, &sn.Status, &wid, &bid2, &sn.LastMovementID, &uc, &sn.CreatedAt, &sn.UpdatedAt); err != nil {
			return nil, err
		}
		sn.BatchID = bid
		sn.WarehouseID = wid
		sn.BinID = bid2
		sn.UnitCost = decPtr(uc)
		out = append(out, sn)
	}
	return out, rows.Err()
}

// MovementHistoryForSerial журнал по серийнику.
func (s *Store) MovementHistoryForSerial(ctx context.Context, tx pgx.Tx, serialID uuid.UUID) ([]models.StockMovement, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, movement_type, document_id, warehouse_id, bin_id, product_id, batch_id, serial_id,
		       qty::text, unit_cost::text, value::text, currency, posted_at, posted_by
		FROM stock_movements WHERE serial_id = $1 ORDER BY posted_at
	`, serialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMovements(rows)
}
