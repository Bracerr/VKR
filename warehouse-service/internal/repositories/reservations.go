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

// CreateReservation создаёт резерв.
func (s *Store) CreateReservation(ctx context.Context, tx pgx.Tx, r *models.Reservation) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO reservations (id, tenant_code, status, warehouse_id, bin_id, product_id, batch_id, serial_id, qty, reason, doc_ref, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, r.ID, r.TenantCode, r.Status, r.WarehouseID, r.BinID, r.ProductID, r.BatchID, r.SerialID, r.Qty.StringFixed(3), r.Reason, r.DocRef, r.ExpiresAt, r.CreatedBy)
	return err
}

// GetReservation резерв.
func (s *Store) GetReservation(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Reservation, error) {
	var r models.Reservation
	var qs string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, status, warehouse_id, bin_id, product_id, batch_id, serial_id, qty::text, reason, doc_ref, expires_at, created_by, created_at, updated_at
		FROM reservations WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&r.ID, &r.TenantCode, &r.Status, &r.WarehouseID, &r.BinID, &r.ProductID, &r.BatchID, &r.SerialID, &qs, &r.Reason, &r.DocRef, &r.ExpiresAt, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	r.Qty, _ = decimal.NewFromString(qs)
	return &r, err
}

// UpdateReservationStatus статус.
func (s *Store) UpdateReservationStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	_, err := s.db(tx).Exec(ctx, `UPDATE reservations SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	return err
}

// ListReservations список.
func (s *Store) ListReservations(ctx context.Context, tx pgx.Tx, tenant string, status *string, whID, productID *uuid.UUID) ([]models.Reservation, error) {
	q := `SELECT id, tenant_code, status, warehouse_id, bin_id, product_id, batch_id, serial_id, qty::text, reason, doc_ref, expires_at, created_by, created_at, updated_at
		FROM reservations WHERE tenant_code = $1`
	args := []interface{}{tenant}
	n := 2
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
	if productID != nil {
		q += ` AND product_id = $` + strconv.Itoa(n)
		args = append(args, *productID)
		n++
	}
	q += ` ORDER BY created_at DESC LIMIT 500`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Reservation
	for rows.Next() {
		var r models.Reservation
		var qs string
		if err := rows.Scan(&r.ID, &r.TenantCode, &r.Status, &r.WarehouseID, &r.BinID, &r.ProductID, &r.BatchID, &r.SerialID, &qs, &r.Reason, &r.DocRef, &r.ExpiresAt, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Qty, _ = decimal.NewFromString(qs)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListActiveExpiredReservations активные резервы с истекшим expires_at.
func (s *Store) ListActiveExpiredReservations(ctx context.Context, tx pgx.Tx, before time.Time) ([]models.Reservation, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, status, warehouse_id, bin_id, product_id, batch_id, serial_id, qty::text, reason, doc_ref, expires_at, created_by, created_at, updated_at
		FROM reservations
		WHERE status = 'ACTIVE' AND expires_at IS NOT NULL AND expires_at < $1
		ORDER BY expires_at
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Reservation
	for rows.Next() {
		var r models.Reservation
		var qs string
		if err := rows.Scan(&r.ID, &r.TenantCode, &r.Status, &r.WarehouseID, &r.BinID, &r.ProductID, &r.BatchID, &r.SerialID, &qs, &r.Reason, &r.DocRef, &r.ExpiresAt, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Qty, _ = decimal.NewFromString(qs)
		out = append(out, r)
	}
	return out, rows.Err()
}
