package repositories

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListRoutings список.
func (s *Store) ListRoutings(ctx context.Context, tx pgx.Tx, tenant string, productID *uuid.UUID, status *string) ([]models.Routing, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, created_at, updated_at
		FROM routings WHERE tenant_code=$1`
	args := []interface{}{tenant}
	n := 2
	if productID != nil {
		q += ` AND product_id=$` + strconv.Itoa(n)
		args = append(args, *productID)
		n++
	}
	if status != nil && *status != "" {
		q += ` AND status=$` + strconv.Itoa(n)
		args = append(args, *status)
	}
	q += ` ORDER BY product_id, version DESC`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Routing
	for rows.Next() {
		var r models.Routing
		if err := rows.Scan(&r.ID, &r.TenantCode, &r.ProductID, &r.Version, &r.Status, &r.SedDocumentID, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetRouting возвращает техкарту.
func (s *Store) GetRouting(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Routing, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, created_at, updated_at
		FROM routings WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var r models.Routing
	if err := row.Scan(&r.ID, &r.TenantCode, &r.ProductID, &r.Version, &r.Status, &r.SedDocumentID, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// FindRoutingBySedDocument находит по документу СЭД.
func (s *Store) FindRoutingBySedDocument(ctx context.Context, tx pgx.Tx, tenant string, sedDocID uuid.UUID) (*models.Routing, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, created_at, updated_at
		FROM routings WHERE tenant_code=$1 AND sed_document_id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, sedDocID)
	var r models.Routing
	if err := row.Scan(&r.ID, &r.TenantCode, &r.ProductID, &r.Version, &r.Status, &r.SedDocumentID, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// NextRoutingVersion следующая версия.
func (s *Store) NextRoutingVersion(ctx context.Context, tx pgx.Tx, tenant string, productID uuid.UUID) (int, error) {
	q := `SELECT COALESCE(MAX(version),0)+1 FROM routings WHERE tenant_code=$1 AND product_id=$2`
	var v int
	err := s.db(tx).QueryRow(ctx, q, tenant, productID).Scan(&v)
	return v, err
}

// CreateRouting создаёт.
func (s *Store) CreateRouting(ctx context.Context, tx pgx.Tx, r *models.Routing) error {
	q := `INSERT INTO routings(id, tenant_code, product_id, version, status, sed_document_id)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, r.ID, r.TenantCode, r.ProductID, r.Version, r.Status, r.SedDocumentID)
	return err
}

// SetRoutingStatusSED статус и sed.
func (s *Store) SetRoutingStatusSED(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string, sedDocID *uuid.UUID) error {
	q := `UPDATE routings SET status=$3, sed_document_id=$4, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status, sedDocID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListRoutingOperations операции маршрута.
func (s *Store) ListRoutingOperations(ctx context.Context, tx pgx.Tx, routingID uuid.UUID) ([]models.RoutingOperation, error) {
	q := `SELECT id, routing_id, op_no, workcenter_id, name,
		CASE WHEN time_per_unit_min IS NULL THEN NULL ELSE time_per_unit_min::text END,
		CASE WHEN setup_time_min IS NULL THEN NULL ELSE setup_time_min::text END,
		qc_required
		FROM routing_operations WHERE routing_id=$1 ORDER BY op_no`
	rows, err := s.db(tx).Query(ctx, q, routingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RoutingOperation
	for rows.Next() {
		var op models.RoutingOperation
		if err := rows.Scan(&op.ID, &op.RoutingID, &op.OpNo, &op.WorkcenterID, &op.Name, &op.TimePerUnitMin, &op.SetupTimeMin, &op.QCRequired); err != nil {
			return nil, err
		}
		out = append(out, op)
	}
	return out, rows.Err()
}

// AddRoutingOperation добавляет операцию.
func (s *Store) AddRoutingOperation(ctx context.Context, tx pgx.Tx, op *models.RoutingOperation) error {
	q := `INSERT INTO routing_operations(id, routing_id, op_no, workcenter_id, name, time_per_unit_min, setup_time_min, qc_required)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	var tpu, stu interface{}
	if op.TimePerUnitMin != nil && *op.TimePerUnitMin != "" {
		tpu = *op.TimePerUnitMin
	}
	if op.SetupTimeMin != nil && *op.SetupTimeMin != "" {
		stu = *op.SetupTimeMin
	}
	_, err := s.db(tx).Exec(ctx, q, op.ID, op.RoutingID, op.OpNo, op.WorkcenterID, op.Name, tpu, stu, op.QCRequired)
	return err
}

// DeleteRoutingOperations удаляет все операции маршрута.
func (s *Store) DeleteRoutingOperations(ctx context.Context, tx pgx.Tx, routingID uuid.UUID) error {
	_, err := s.db(tx).Exec(ctx, `DELETE FROM routing_operations WHERE routing_id=$1`, routingID)
	return err
}
