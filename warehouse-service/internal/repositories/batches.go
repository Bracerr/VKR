package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// CreateBatch создаёт партию.
func (s *Store) CreateBatch(ctx context.Context, tx pgx.Tx, b *models.Batch) error {
	var uc *string
	if b.UnitCost != nil {
		v := b.UnitCost.StringFixed(4)
		uc = &v
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO batches (id, tenant_code, product_id, series, manufactured_at, expires_at, unit_cost, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, b.ID, b.TenantCode, b.ProductID, b.Series, b.ManufacturedAt, b.ExpiresAt, uc, b.Currency)
	return err
}

// GetBatchByProductSeries партия по товару и серии.
func (s *Store) GetBatchByProductSeries(ctx context.Context, tx pgx.Tx, productID uuid.UUID, series string) (*models.Batch, error) {
	var b models.Batch
	var uc *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, product_id, series, manufactured_at, expires_at, unit_cost::text, currency, created_at
		FROM batches WHERE product_id = $1 AND series = $2
	`, productID, series).Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Series, &b.ManufacturedAt, &b.ExpiresAt, &uc, &b.Currency, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	b.UnitCost = decPtr(uc)
	return &b, err
}

// GetBatchByID партия.
func (s *Store) GetBatchByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Batch, error) {
	var b models.Batch
	var uc *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, product_id, series, manufactured_at, expires_at, unit_cost::text, currency, created_at
		FROM batches WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Series, &b.ManufacturedAt, &b.ExpiresAt, &uc, &b.Currency, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	b.UnitCost = decPtr(uc)
	return &b, err
}

// ListBatchesForFEFO партии с остатком на складе+ячейке для FEFO.
func (s *Store) ListBatchesForFEFO(ctx context.Context, tx pgx.Tx, tenant string, warehouseID, binID, productID uuid.UUID) ([]struct {
	BatchID   uuid.UUID
	ExpiresAt *time.Time
	QtyAvail  decimal.Decimal
}, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT sb.batch_id, b.expires_at, (sb.quantity - sb.reserved_qty)::text
		FROM stock_balances sb
		JOIN batches b ON b.id = sb.batch_id
		WHERE sb.warehouse_id = $1 AND sb.bin_id = $2 AND sb.product_id = $3
		  AND b.tenant_code = $4
		  AND sb.batch_id <> wh_nil_batch()
		  AND (sb.quantity - sb.reserved_qty) > 0
		ORDER BY b.expires_at NULLS LAST, b.id
	`, warehouseID, binID, productID, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		BatchID   uuid.UUID
		ExpiresAt *time.Time
		QtyAvail  decimal.Decimal
	}
	for rows.Next() {
		var bid uuid.UUID
		var exp *time.Time
		var qs string
		if err := rows.Scan(&bid, &exp, &qs); err != nil {
			return nil, err
		}
		q, _ := decimal.NewFromString(qs)
		out = append(out, struct {
			BatchID   uuid.UUID
			ExpiresAt *time.Time
			QtyAvail  decimal.Decimal
		}{BatchID: bid, ExpiresAt: exp, QtyAvail: q})
	}
	return out, rows.Err()
}

// ListExpiringBatches скоро истекающие. tenant "" — все тенанты.
func (s *Store) ListExpiringBatches(ctx context.Context, tx pgx.Tx, tenant string, before time.Time) ([]models.Batch, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, product_id, series, manufactured_at, expires_at, unit_cost::text, currency, created_at
		FROM batches WHERE ($1 = '' OR tenant_code = $1) AND expires_at IS NOT NULL AND expires_at <= $2::date
		ORDER BY expires_at
	`, tenant, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Batch
	for rows.Next() {
		var b models.Batch
		var uc *string
		if err := rows.Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Series, &b.ManufacturedAt, &b.ExpiresAt, &uc, &b.Currency, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.UnitCost = decPtr(uc)
		out = append(out, b)
	}
	return out, rows.Err()
}
