package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/procurement-service/internal/models"
)

func (s *Store) CreatePO(ctx context.Context, tx pgx.Tx, po *models.PurchaseOrder) error {
	q := `INSERT INTO purchase_orders(id, tenant_code, number, supplier_id, status, created_by_sub, currency, expected_at, sed_document_id, source_pr_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.db(tx).Exec(ctx, q, po.ID, po.TenantCode, po.Number, po.SupplierID, po.Status, po.CreatedBySub, po.Currency, po.ExpectedAt, po.SedDocumentID, po.SourcePRID)
	return err
}

func (s *Store) GetPO(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.PurchaseOrder, error) {
	q := `SELECT id, tenant_code, number, supplier_id, status, created_by_sub, currency, expected_at, sed_document_id, source_pr_id, created_at, updated_at
		FROM purchase_orders WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var po models.PurchaseOrder
	if err := row.Scan(&po.ID, &po.TenantCode, &po.Number, &po.SupplierID, &po.Status, &po.CreatedBySub, &po.Currency, &po.ExpectedAt, &po.SedDocumentID, &po.SourcePRID, &po.CreatedAt, &po.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (s *Store) LockPO(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.PurchaseOrder, error) {
	q := `SELECT id, tenant_code, number, supplier_id, status, created_by_sub, currency, expected_at, sed_document_id, source_pr_id, created_at, updated_at
		FROM purchase_orders WHERE tenant_code=$1 AND id=$2 FOR UPDATE`
	row := tx.QueryRow(ctx, q, tenant, id)
	var po models.PurchaseOrder
	if err := row.Scan(&po.ID, &po.TenantCode, &po.Number, &po.SupplierID, &po.Status, &po.CreatedBySub, &po.Currency, &po.ExpectedAt, &po.SedDocumentID, &po.SourcePRID, &po.CreatedAt, &po.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (s *Store) ListPO(ctx context.Context, tx pgx.Tx, tenant string, status *string) ([]models.PurchaseOrder, error) {
	q := `SELECT id, tenant_code, number, supplier_id, status, created_by_sub, currency, expected_at, sed_document_id, source_pr_id, created_at, updated_at
		FROM purchase_orders WHERE tenant_code=$1`
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
	var out []models.PurchaseOrder
	for rows.Next() {
		var po models.PurchaseOrder
		if err := rows.Scan(&po.ID, &po.TenantCode, &po.Number, &po.SupplierID, &po.Status, &po.CreatedBySub, &po.Currency, &po.ExpectedAt, &po.SedDocumentID, &po.SourcePRID, &po.CreatedAt, &po.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, po)
	}
	return out, rows.Err()
}

func (s *Store) UpdatePOStatusSed(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string, sedDocID *uuid.UUID) error {
	q := `UPDATE purchase_orders SET status=$3, sed_document_id=$4, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status, sedDocID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) AddPOLine(ctx context.Context, tx pgx.Tx, ln *models.PurchaseOrderLine) error {
	q := `INSERT INTO purchase_order_lines(id, po_id, line_no, product_id, qty_ordered, qty_received, price, vat_rate, target_warehouse_id, target_bin_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.db(tx).Exec(ctx, q, ln.ID, ln.POID, ln.LineNo, ln.ProductID, ln.QtyOrdered, ln.QtyReceived, ln.Price, ln.VATRate, ln.TargetWarehouseID, ln.TargetBinID)
	return err
}

func (s *Store) ListPOLines(ctx context.Context, tx pgx.Tx, poID uuid.UUID) ([]models.PurchaseOrderLine, error) {
	q := `SELECT id, po_id, line_no, product_id, qty_ordered::text, qty_received::text, price::text, vat_rate::text, target_warehouse_id, target_bin_id
		FROM purchase_order_lines WHERE po_id=$1 ORDER BY line_no`
	rows, err := s.db(tx).Query(ctx, q, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PurchaseOrderLine
	for rows.Next() {
		var ln models.PurchaseOrderLine
		if err := rows.Scan(&ln.ID, &ln.POID, &ln.LineNo, &ln.ProductID, &ln.QtyOrdered, &ln.QtyReceived, &ln.Price, &ln.VATRate, &ln.TargetWarehouseID, &ln.TargetBinID); err != nil {
			return nil, err
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}

func (s *Store) FindPOBySedDocument(ctx context.Context, tx pgx.Tx, tenant string, sedDocID uuid.UUID) (*models.PurchaseOrder, error) {
	q := `SELECT id, tenant_code, number, supplier_id, status, created_by_sub, currency, expected_at, sed_document_id, source_pr_id, created_at, updated_at
		FROM purchase_orders WHERE tenant_code=$1 AND sed_document_id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, sedDocID)
	var po models.PurchaseOrder
	if err := row.Scan(&po.ID, &po.TenantCode, &po.Number, &po.SupplierID, &po.Status, &po.CreatedBySub, &po.Currency, &po.ExpectedAt, &po.SedDocumentID, &po.SourcePRID, &po.CreatedAt, &po.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (s *Store) InsertReceipt(ctx context.Context, tx pgx.Tx, r *models.Receipt) error {
	q := `INSERT INTO receipts(id, tenant_code, po_id, warehouse_document_id, status, posted_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, r.ID, r.TenantCode, r.POID, r.WarehouseDocumentID, r.Status, r.PostedAt)
	return err
}

