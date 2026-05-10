package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/procurement-service/internal/models"
)

func (s *Store) CreatePR(ctx context.Context, tx pgx.Tx, pr *models.PurchaseRequest) error {
	q := `INSERT INTO purchase_requests(id, tenant_code, number, status, created_by_sub, needed_by, note, sed_document_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.db(tx).Exec(ctx, q, pr.ID, pr.TenantCode, pr.Number, pr.Status, pr.CreatedBySub, pr.NeededBy, pr.Note, pr.SedDocumentID)
	return err
}

func (s *Store) GetPR(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.PurchaseRequest, error) {
	q := `SELECT id, tenant_code, number, status, created_by_sub, needed_by, note, sed_document_id, created_at, updated_at
		FROM purchase_requests WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var pr models.PurchaseRequest
	if err := row.Scan(&pr.ID, &pr.TenantCode, &pr.Number, &pr.Status, &pr.CreatedBySub, &pr.NeededBy, &pr.Note, &pr.SedDocumentID, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func (s *Store) LockPR(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.PurchaseRequest, error) {
	q := `SELECT id, tenant_code, number, status, created_by_sub, needed_by, note, sed_document_id, created_at, updated_at
		FROM purchase_requests WHERE tenant_code=$1 AND id=$2 FOR UPDATE`
	row := tx.QueryRow(ctx, q, tenant, id)
	var pr models.PurchaseRequest
	if err := row.Scan(&pr.ID, &pr.TenantCode, &pr.Number, &pr.Status, &pr.CreatedBySub, &pr.NeededBy, &pr.Note, &pr.SedDocumentID, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func (s *Store) ListPR(ctx context.Context, tx pgx.Tx, tenant string, status *string) ([]models.PurchaseRequest, error) {
	q := `SELECT id, tenant_code, number, status, created_by_sub, needed_by, note, sed_document_id, created_at, updated_at
		FROM purchase_requests WHERE tenant_code=$1`
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
	var out []models.PurchaseRequest
	for rows.Next() {
		var pr models.PurchaseRequest
		if err := rows.Scan(&pr.ID, &pr.TenantCode, &pr.Number, &pr.Status, &pr.CreatedBySub, &pr.NeededBy, &pr.Note, &pr.SedDocumentID, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, pr)
	}
	return out, rows.Err()
}

func (s *Store) UpdatePRStatusSed(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string, sedDocID *uuid.UUID) error {
	q := `UPDATE purchase_requests SET status=$3, sed_document_id=$4, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status, sedDocID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) AddPRLine(ctx context.Context, tx pgx.Tx, ln *models.PurchaseRequestLine) error {
	q := `INSERT INTO purchase_request_lines(id, pr_id, line_no, product_id, qty, uom, target_warehouse_id, target_bin_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := s.db(tx).Exec(ctx, q, ln.ID, ln.PRID, ln.LineNo, ln.ProductID, ln.Qty, ln.UOM, ln.TargetWarehouseID, ln.TargetBinID, ln.Note)
	return err
}

func (s *Store) ListPRLines(ctx context.Context, tx pgx.Tx, prID uuid.UUID) ([]models.PurchaseRequestLine, error) {
	q := `SELECT id, pr_id, line_no, product_id, qty::text, uom, target_warehouse_id, target_bin_id, note
		FROM purchase_request_lines WHERE pr_id=$1 ORDER BY line_no`
	rows, err := s.db(tx).Query(ctx, q, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PurchaseRequestLine
	for rows.Next() {
		var ln models.PurchaseRequestLine
		if err := rows.Scan(&ln.ID, &ln.PRID, &ln.LineNo, &ln.ProductID, &ln.Qty, &ln.UOM, &ln.TargetWarehouseID, &ln.TargetBinID, &ln.Note); err != nil {
			return nil, err
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}

func (s *Store) FindPRBySedDocument(ctx context.Context, tx pgx.Tx, tenant string, sedDocID uuid.UUID) (*models.PurchaseRequest, error) {
	q := `SELECT id, tenant_code, number, status, created_by_sub, needed_by, note, sed_document_id, created_at, updated_at
		FROM purchase_requests WHERE tenant_code=$1 AND sed_document_id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, sedDocID)
	var pr models.PurchaseRequest
	if err := row.Scan(&pr.ID, &pr.TenantCode, &pr.Number, &pr.Status, &pr.CreatedBySub, &pr.NeededBy, &pr.Note, &pr.SedDocumentID, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

