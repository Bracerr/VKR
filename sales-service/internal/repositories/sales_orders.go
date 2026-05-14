package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/sales-service/internal/models"
)

func (s *Store) CreateSO(ctx context.Context, tx pgx.Tx, so *models.SalesOrder) error {
	q := `INSERT INTO sales_orders(id, tenant_code, number, status, customer_id, created_by_sub, ship_from_warehouse_id, ship_from_bin_id, note, sed_document_id, reservations)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := s.db(tx).Exec(ctx, q, so.ID, so.TenantCode, so.Number, so.Status, so.CustomerID, so.CreatedBySub, so.ShipFromWarehouseID, so.ShipFromBinID, so.Note, so.SedDocumentID, so.Reservations)
	return err
}

func (s *Store) GetSO(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.SalesOrder, error) {
	q := `SELECT id, tenant_code, number, status, customer_id, created_by_sub, ship_from_warehouse_id, ship_from_bin_id, note, sed_document_id, reservations, created_at, updated_at
		FROM sales_orders WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var so models.SalesOrder
	if err := row.Scan(&so.ID, &so.TenantCode, &so.Number, &so.Status, &so.CustomerID, &so.CreatedBySub, &so.ShipFromWarehouseID, &so.ShipFromBinID, &so.Note, &so.SedDocumentID, &so.Reservations, &so.CreatedAt, &so.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &so, nil
}

func (s *Store) LockSO(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.SalesOrder, error) {
	q := `SELECT id, tenant_code, number, status, customer_id, created_by_sub, ship_from_warehouse_id, ship_from_bin_id, note, sed_document_id, reservations, created_at, updated_at
		FROM sales_orders WHERE tenant_code=$1 AND id=$2 FOR UPDATE`
	row := tx.QueryRow(ctx, q, tenant, id)
	var so models.SalesOrder
	if err := row.Scan(&so.ID, &so.TenantCode, &so.Number, &so.Status, &so.CustomerID, &so.CreatedBySub, &so.ShipFromWarehouseID, &so.ShipFromBinID, &so.Note, &so.SedDocumentID, &so.Reservations, &so.CreatedAt, &so.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &so, nil
}

func (s *Store) ListSO(ctx context.Context, tx pgx.Tx, tenant string, status *string) ([]models.SalesOrder, error) {
	q := `SELECT id, tenant_code, number, status, customer_id, created_by_sub, ship_from_warehouse_id, ship_from_bin_id, note, sed_document_id, reservations, created_at, updated_at
		FROM sales_orders WHERE tenant_code=$1`
	args := []any{tenant}
	if status != nil && *status != "" {
		q += ` AND status=$2`
		args = append(args, *status)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SalesOrder
	for rows.Next() {
		var so models.SalesOrder
		if err := rows.Scan(&so.ID, &so.TenantCode, &so.Number, &so.Status, &so.CustomerID, &so.CreatedBySub, &so.ShipFromWarehouseID, &so.ShipFromBinID, &so.Note, &so.SedDocumentID, &so.Reservations, &so.CreatedAt, &so.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, so)
	}
	return out, rows.Err()
}

func (s *Store) UpdateSOStatusSedAndReservations(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string, sedDocID *uuid.UUID, reservations []byte) error {
	q := `UPDATE sales_orders SET status=$3, sed_document_id=$4, reservations=$5, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status, sedDocID, reservations)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) FindSOBySedDocument(ctx context.Context, tx pgx.Tx, tenant string, sedDocID uuid.UUID) (*models.SalesOrder, error) {
	q := `SELECT id, tenant_code, number, status, customer_id, created_by_sub, ship_from_warehouse_id, ship_from_bin_id, note, sed_document_id, reservations, created_at, updated_at
		FROM sales_orders WHERE tenant_code=$1 AND sed_document_id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, sedDocID)
	var so models.SalesOrder
	if err := row.Scan(&so.ID, &so.TenantCode, &so.Number, &so.Status, &so.CustomerID, &so.CreatedBySub, &so.ShipFromWarehouseID, &so.ShipFromBinID, &so.Note, &so.SedDocumentID, &so.Reservations, &so.CreatedAt, &so.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &so, nil
}

func (s *Store) AddSOLine(ctx context.Context, tx pgx.Tx, ln *models.SalesOrderLine) error {
	q := `INSERT INTO sales_order_lines(id, so_id, line_no, product_id, qty, uom, reserved_qty, shipped_qty, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := s.db(tx).Exec(ctx, q, ln.ID, ln.SOID, ln.LineNo, ln.ProductID, ln.Qty, ln.UOM, ln.ReservedQty, ln.ShippedQty, ln.Note)
	return err
}

func (s *Store) ListSOLines(ctx context.Context, tx pgx.Tx, soID uuid.UUID) ([]models.SalesOrderLine, error) {
	q := `SELECT id, so_id, line_no, product_id, qty::text, uom, reserved_qty::text, shipped_qty::text, note
		FROM sales_order_lines WHERE so_id=$1 ORDER BY line_no`
	rows, err := s.db(tx).Query(ctx, q, soID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SalesOrderLine
	for rows.Next() {
		var ln models.SalesOrderLine
		if err := rows.Scan(&ln.ID, &ln.SOID, &ln.LineNo, &ln.ProductID, &ln.Qty, &ln.UOM, &ln.ReservedQty, &ln.ShippedQty, &ln.Note); err != nil {
			return nil, err
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}

func (s *Store) InsertShipment(ctx context.Context, tx pgx.Tx, sh *models.Shipment) error {
	q := `INSERT INTO shipments(id, tenant_code, so_id, warehouse_document_id, status, posted_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, sh.ID, sh.TenantCode, sh.SOID, sh.WarehouseDocumentID, sh.Status, sh.PostedAt)
	return err
}

func (s *Store) GetShipmentBySO(ctx context.Context, tx pgx.Tx, tenant string, soID uuid.UUID) (*models.Shipment, error) {
	q := `SELECT id, tenant_code, so_id, warehouse_document_id, status, posted_at
		FROM shipments WHERE tenant_code=$1 AND so_id=$2 ORDER BY posted_at DESC LIMIT 1`
	row := s.db(tx).QueryRow(ctx, q, tenant, soID)
	var sh models.Shipment
	if err := row.Scan(&sh.ID, &sh.TenantCode, &sh.SOID, &sh.WarehouseDocumentID, &sh.Status, &sh.PostedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sh, nil
}

