package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListProductionOrders список.
func (s *Store) ListProductionOrders(ctx context.Context, tx pgx.Tx, tenant string, status *string) ([]models.ProductionOrder, error) {
	q := `SELECT id, tenant_code, code, product_id, qty_planned::text, qty_done::text, qty_scrap::text, status,
		bom_id, routing_id, warehouse_id, default_bin_id, reservations, warehouse_receipt_doc_id,
		start_plan, finish_plan, start_fact, finish_fact, created_at, updated_at
		FROM production_orders WHERE tenant_code=$1`
	args := []interface{}{tenant}
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
	return scanProductionOrders(rows)
}

func scanProductionOrders(rows pgx.Rows) ([]models.ProductionOrder, error) {
	var out []models.ProductionOrder
	for rows.Next() {
		var o models.ProductionOrder
		if err := rows.Scan(&o.ID, &o.TenantCode, &o.Code, &o.ProductID, &o.QtyPlanned, &o.QtyDone, &o.QtyScrap, &o.Status,
			&o.BomID, &o.RoutingID, &o.WarehouseID, &o.DefaultBinID, &o.Reservations, &o.WarehouseReceiptDocID,
			&o.StartPlan, &o.FinishPlan, &o.StartFact, &o.FinishFact, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// GetProductionOrder заказ.
func (s *Store) GetProductionOrder(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.ProductionOrder, error) {
	q := `SELECT id, tenant_code, code, product_id, qty_planned::text, qty_done::text, qty_scrap::text, status,
		bom_id, routing_id, warehouse_id, default_bin_id, reservations, warehouse_receipt_doc_id,
		start_plan, finish_plan, start_fact, finish_fact, created_at, updated_at
		FROM production_orders WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var o models.ProductionOrder
	if err := row.Scan(&o.ID, &o.TenantCode, &o.Code, &o.ProductID, &o.QtyPlanned, &o.QtyDone, &o.QtyScrap, &o.Status,
		&o.BomID, &o.RoutingID, &o.WarehouseID, &o.DefaultBinID, &o.Reservations, &o.WarehouseReceiptDocID,
		&o.StartPlan, &o.FinishPlan, &o.StartFact, &o.FinishFact, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// LockProductionOrder FOR UPDATE.
func (s *Store) LockProductionOrder(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.ProductionOrder, error) {
	q := `SELECT id, tenant_code, code, product_id, qty_planned::text, qty_done::text, qty_scrap::text, status,
		bom_id, routing_id, warehouse_id, default_bin_id, reservations, warehouse_receipt_doc_id,
		start_plan, finish_plan, start_fact, finish_fact, created_at, updated_at
		FROM production_orders WHERE tenant_code=$1 AND id=$2 FOR UPDATE`
	row := tx.QueryRow(ctx, q, tenant, id)
	var o models.ProductionOrder
	if err := row.Scan(&o.ID, &o.TenantCode, &o.Code, &o.ProductID, &o.QtyPlanned, &o.QtyDone, &o.QtyScrap, &o.Status,
		&o.BomID, &o.RoutingID, &o.WarehouseID, &o.DefaultBinID, &o.Reservations, &o.WarehouseReceiptDocID,
		&o.StartPlan, &o.FinishPlan, &o.StartFact, &o.FinishFact, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// CreateProductionOrder создаёт.
func (s *Store) CreateProductionOrder(ctx context.Context, tx pgx.Tx, o *models.ProductionOrder) error {
	q := `INSERT INTO production_orders(id, tenant_code, code, product_id, qty_planned, qty_done, qty_scrap, status,
		bom_id, routing_id, warehouse_id, default_bin_id, reservations, warehouse_receipt_doc_id,
		start_plan, finish_plan)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err := s.db(tx).Exec(ctx, q,
		o.ID, o.TenantCode, o.Code, o.ProductID, o.QtyPlanned, o.QtyDone, o.QtyScrap, o.Status,
		o.BomID, o.RoutingID, o.WarehouseID, o.DefaultBinID, o.Reservations, o.WarehouseReceiptDocID,
		o.StartPlan, o.FinishPlan)
	return err
}

// UpdateProductionOrderReservations обновляет резервы JSON.
func (s *Store) UpdateProductionOrderReservations(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, reservations []byte, status string) error {
	q := `UPDATE production_orders SET reservations=$3, status=$4, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, reservations, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdateProductionOrderStatus только статус.
func (s *Store) UpdateProductionOrderStatus(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string) error {
	q := `UPDATE production_orders SET status=$3, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdateProductionOrderProgress qty и статус.
func (s *Store) UpdateProductionOrderProgress(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, qtyDone, qtyScrap, status string, startFact, finishFact interface{}) error {
	q := `UPDATE production_orders SET qty_done=$3, qty_scrap=$4, status=$5,
		start_fact=COALESCE(start_fact, $6::timestamptz), finish_fact=$7::timestamptz, updated_at=now()
		WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, qtyDone, qtyScrap, status, startFact, finishFact)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetProductionOrderReceiptReceipt сохраняет приход GP.
func (s *Store) SetProductionOrderReceipt(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, receiptDocID uuid.UUID, status string, finishFact interface{}) error {
	q := `UPDATE production_orders SET warehouse_receipt_doc_id=$3, status=$4, finish_fact=$5, updated_at=now()
		WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, receiptDocID, status, finishFact)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
