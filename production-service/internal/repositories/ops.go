package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListOrderOperations операции заказа.
func (s *Store) ListOrderOperations(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) ([]models.OrderOperation, error) {
	q := `SELECT id, order_id, op_no, workcenter_id, name, qty_planned::text, qty_good::text, qty_scrap::text, status, started_at, finished_at
		FROM production_order_operations WHERE order_id=$1 ORDER BY op_no`
	rows, err := s.db(tx).Query(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.OrderOperation
	for rows.Next() {
		var op models.OrderOperation
		if err := rows.Scan(&op.ID, &op.OrderID, &op.OpNo, &op.WorkcenterID, &op.Name, &op.QtyPlanned, &op.QtyGood, &op.QtyScrap, &op.Status, &op.StartedAt, &op.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, op)
	}
	return out, rows.Err()
}

// GetOrderOperationByID только по id операции (для shift_tasks).
func (s *Store) GetOrderOperationByID(ctx context.Context, tx pgx.Tx, tenant string, opID uuid.UUID) (*models.OrderOperation, error) {
	q := `SELECT oo.id, oo.order_id, oo.op_no, oo.workcenter_id, oo.name, oo.qty_planned::text, oo.qty_good::text, oo.qty_scrap::text, oo.status, oo.started_at, oo.finished_at
		FROM production_order_operations oo
		JOIN production_orders o ON o.id = oo.order_id
		WHERE oo.id=$1 AND o.tenant_code=$2`
	row := s.db(tx).QueryRow(ctx, q, opID, tenant)
	var op models.OrderOperation
	if err := row.Scan(&op.ID, &op.OrderID, &op.OpNo, &op.WorkcenterID, &op.Name, &op.QtyPlanned, &op.QtyGood, &op.QtyScrap, &op.Status, &op.StartedAt, &op.FinishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &op, nil
}

// GetOrderOperation по id.
func (s *Store) GetOrderOperation(ctx context.Context, tx pgx.Tx, orderID, opID uuid.UUID) (*models.OrderOperation, error) {
	q := `SELECT id, order_id, op_no, workcenter_id, name, qty_planned::text, qty_good::text, qty_scrap::text, status, started_at, finished_at
		FROM production_order_operations WHERE order_id=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, orderID, opID)
	var op models.OrderOperation
	if err := row.Scan(&op.ID, &op.OrderID, &op.OpNo, &op.WorkcenterID, &op.Name, &op.QtyPlanned, &op.QtyGood, &op.QtyScrap, &op.Status, &op.StartedAt, &op.FinishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &op, nil
}

// LockOrderOperation FOR UPDATE.
func (s *Store) LockOrderOperation(ctx context.Context, tx pgx.Tx, orderID, opID uuid.UUID) (*models.OrderOperation, error) {
	q := `SELECT id, order_id, op_no, workcenter_id, name, qty_planned::text, qty_good::text, qty_scrap::text, status, started_at, finished_at
		FROM production_order_operations WHERE order_id=$1 AND id=$2 FOR UPDATE`
	row := tx.QueryRow(ctx, q, orderID, opID)
	var op models.OrderOperation
	if err := row.Scan(&op.ID, &op.OrderID, &op.OpNo, &op.WorkcenterID, &op.Name, &op.QtyPlanned, &op.QtyGood, &op.QtyScrap, &op.Status, &op.StartedAt, &op.FinishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &op, nil
}

// InsertOrderOperation snapshot.
func (s *Store) InsertOrderOperation(ctx context.Context, tx pgx.Tx, op *models.OrderOperation) error {
	q := `INSERT INTO production_order_operations(id, order_id, op_no, workcenter_id, name, qty_planned, qty_good, qty_scrap, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := s.db(tx).Exec(ctx, q, op.ID, op.OrderID, op.OpNo, op.WorkcenterID, op.Name, op.QtyPlanned, op.QtyGood, op.QtyScrap, op.Status)
	return err
}

// UpdateOrderOperation обновляет числа и статус.
func (s *Store) UpdateOrderOperation(ctx context.Context, tx pgx.Tx, op *models.OrderOperation) error {
	q := `UPDATE production_order_operations SET qty_good=$2, qty_scrap=$3, status=$4, started_at=$5, finished_at=$6 WHERE id=$1`
	tag, err := s.db(tx).Exec(ctx, q, op.ID, op.QtyGood, op.QtyScrap, op.Status, op.StartedAt, op.FinishedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// InsertProductionReport отчёт.
func (s *Store) InsertProductionReport(ctx context.Context, tx pgx.Tx, r *models.ProductionReport) error {
	q := `INSERT INTO production_reports(id, tenant_code, order_operation_id, reported_by_sub, qty_good, qty_scrap, scrap_reason_code, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.db(tx).Exec(ctx, q, r.ID, r.TenantCode, r.OrderOperationID, r.ReportedBySub, r.QtyGood, r.QtyScrap, r.ScrapReasonCode, r.Note)
	return err
}
