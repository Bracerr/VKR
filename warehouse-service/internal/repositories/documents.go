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

// CreateDocument создаёт документ.
func (s *Store) CreateDocument(ctx context.Context, tx pgx.Tx, d *models.Document) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO documents (id, tenant_code, doc_type, number, status, warehouse_from_id, warehouse_to_id, period_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, d.ID, d.TenantCode, d.DocType, d.Number, d.Status, d.WarehouseFromID, d.WarehouseToID, d.PeriodAt, d.CreatedBy)
	return err
}

// UpdateDocumentStatus статус.
func (s *Store) UpdateDocumentStatus(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string) error {
	_, err := s.db(tx).Exec(ctx, `UPDATE documents SET status = $3 WHERE id = $1 AND tenant_code = $2`, id, tenant, status)
	return err
}

// GetDocument документ.
func (s *Store) GetDocument(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Document, error) {
	var d models.Document
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, doc_type, number, status, warehouse_from_id, warehouse_to_id, period_at, created_by, created_at
		FROM documents WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&d.ID, &d.TenantCode, &d.DocType, &d.Number, &d.Status, &d.WarehouseFromID, &d.WarehouseToID, &d.PeriodAt, &d.CreatedBy, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &d, err
}

// InsertDocumentLine строка документа.
func (s *Store) InsertDocumentLine(ctx context.Context, tx pgx.Tx, docID, productID uuid.UUID, batchID *uuid.UUID, serialID *uuid.UUID, qty decimal.Decimal, unitCost *decimal.Decimal) error {
	bid := models.NilBatchID
	if batchID != nil {
		bid = *batchID
	}
	var uc *string
	if unitCost != nil {
		t := unitCost.StringFixed(4)
		uc = &t
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO document_lines (id, document_id, product_id, batch_id, serial_id, qty, unit_cost)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, uuid.New(), docID, productID, bid, serialID, qty.StringFixed(3), uc)
	return err
}

// InsertInventoryLine строка инвентаризации.
func (s *Store) InsertInventoryLine(ctx context.Context, tx pgx.Tx, docID, whID, binID, productID, batchID uuid.UUID, serialID *uuid.UUID, expected decimal.Decimal, counted *decimal.Decimal) error {
	var cnt *string
	if counted != nil {
		t := counted.StringFixed(3)
		cnt = &t
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO inventory_lines (id, document_id, warehouse_id, bin_id, product_id, batch_id, serial_id, expected_qty, counted_qty)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, uuid.New(), docID, whID, binID, productID, batchID, serialID, expected.StringFixed(3), cnt)
	return err
}

// ListInventoryLines строки инвентаризации (с проверкой тенанта).
func (s *Store) ListInventoryLines(ctx context.Context, tx pgx.Tx, tenant string, docID uuid.UUID) ([]struct {
	ID, WhID, BinID, ProductID, BatchID uuid.UUID
	SerialID                             *uuid.UUID
	Expected                             decimal.Decimal
	Counted                              *decimal.Decimal
}, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT il.id, il.warehouse_id, il.bin_id, il.product_id, il.batch_id, il.serial_id, il.expected_qty::text, il.counted_qty::text
		FROM inventory_lines il
		JOIN documents d ON d.id = il.document_id
		WHERE il.document_id = $1 AND d.tenant_code = $2
	`, docID, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, WhID, BinID, ProductID, BatchID uuid.UUID
		SerialID                             *uuid.UUID
		Expected                             decimal.Decimal
		Counted                              *decimal.Decimal
	}
	for rows.Next() {
		var r struct {
			ID, WhID, BinID, ProductID, BatchID uuid.UUID
			SerialID                             *uuid.UUID
			Expected                             decimal.Decimal
			Counted                              *decimal.Decimal
		}
		var expS string
		var cntS *string
		if err := rows.Scan(&r.ID, &r.WhID, &r.BinID, &r.ProductID, &r.BatchID, &r.SerialID, &expS, &cntS); err != nil {
			return nil, err
		}
		r.Expected, _ = decimal.NewFromString(expS)
		if cntS != nil {
			c, _ := decimal.NewFromString(*cntS)
			r.Counted = &c
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateInventoryLineCounted обновить факт (только строка документа тенанта).
func (s *Store) UpdateInventoryLineCounted(ctx context.Context, tx pgx.Tx, tenant string, lineID uuid.UUID, counted decimal.Decimal) error {
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE inventory_lines il SET counted_qty = $2
		FROM documents d
		WHERE il.id = $1 AND il.document_id = d.id AND d.tenant_code = $3
	`, lineID, counted.StringFixed(3), tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListMovements журнал.
func (s *Store) ListMovements(ctx context.Context, tx pgx.Tx, tenant string, from, to time.Time, whID, productID *uuid.UUID, movType *string, limit int) ([]models.StockMovement, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	q := `
		SELECT id, tenant_code, movement_type, document_id, warehouse_id, bin_id, product_id, batch_id, serial_id,
		       qty::text, unit_cost::text, value::text, currency, posted_at, posted_by
		FROM stock_movements WHERE tenant_code = $1 AND posted_at >= $2 AND posted_at < $3`
	args := []interface{}{tenant, from, to}
	n := 4
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
	if movType != nil {
		q += ` AND movement_type = $` + strconv.Itoa(n)
		args = append(args, *movType)
		n++
	}
	q += ` ORDER BY posted_at DESC LIMIT ` + strconv.Itoa(limit)
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanMovements(rows)
}
