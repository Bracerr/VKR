package repositories

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListBOMs фильтр по product/status.
func (s *Store) ListBOMs(ctx context.Context, tx pgx.Tx, tenant string, productID *uuid.UUID, status *string) ([]models.BOM, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, valid_from, valid_to, created_at, updated_at
		FROM boms WHERE tenant_code=$1`
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
	var out []models.BOM
	for rows.Next() {
		var b models.BOM
		if err := rows.Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Version, &b.Status, &b.SedDocumentID, &b.ValidFrom, &b.ValidTo, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBOM возвращает BOM.
func (s *Store) GetBOM(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.BOM, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, valid_from, valid_to, created_at, updated_at
		FROM boms WHERE tenant_code=$1 AND id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, id)
	var b models.BOM
	if err := row.Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Version, &b.Status, &b.SedDocumentID, &b.ValidFrom, &b.ValidTo, &b.CreatedAt, &b.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

// FindBOMBySedDocument находит по документу СЭД.
func (s *Store) FindBOMBySedDocument(ctx context.Context, tx pgx.Tx, tenant string, sedDocID uuid.UUID) (*models.BOM, error) {
	q := `SELECT id, tenant_code, product_id, version, status, sed_document_id, valid_from, valid_to, created_at, updated_at
		FROM boms WHERE tenant_code=$1 AND sed_document_id=$2`
	row := s.db(tx).QueryRow(ctx, q, tenant, sedDocID)
	var b models.BOM
	if err := row.Scan(&b.ID, &b.TenantCode, &b.ProductID, &b.Version, &b.Status, &b.SedDocumentID, &b.ValidFrom, &b.ValidTo, &b.CreatedAt, &b.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

// NextBOMVersion следующая версия для продукта.
func (s *Store) NextBOMVersion(ctx context.Context, tx pgx.Tx, tenant string, productID uuid.UUID) (int, error) {
	q := `SELECT COALESCE(MAX(version),0)+1 FROM boms WHERE tenant_code=$1 AND product_id=$2`
	var v int
	err := s.db(tx).QueryRow(ctx, q, tenant, productID).Scan(&v)
	return v, err
}

// CreateBOM создаёт.
func (s *Store) CreateBOM(ctx context.Context, tx pgx.Tx, b *models.BOM) error {
	q := `INSERT INTO boms(id, tenant_code, product_id, version, status, sed_document_id, valid_from, valid_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.db(tx).Exec(ctx, q, b.ID, b.TenantCode, b.ProductID, b.Version, b.Status, b.SedDocumentID, b.ValidFrom, b.ValidTo)
	return err
}

// UpdateBOM обновляет черновик.
func (s *Store) UpdateBOM(ctx context.Context, tx pgx.Tx, b *models.BOM) error {
	q := `UPDATE boms SET valid_from=$2, valid_to=$3, updated_at=now() WHERE tenant_code=$4 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, b.ID, b.ValidFrom, b.ValidTo, b.TenantCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetBOMStatusSED меняет статус и sed_document_id.
func (s *Store) SetBOMStatusSED(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID, status string, sedDocID *uuid.UUID) error {
	q := `UPDATE boms SET status=$3, sed_document_id=$4, updated_at=now() WHERE tenant_code=$2 AND id=$1`
	tag, err := s.db(tx).Exec(ctx, q, id, tenant, status, sedDocID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListBOMLines строки BOM.
func (s *Store) ListBOMLines(ctx context.Context, tx pgx.Tx, bomID uuid.UUID) ([]models.BOMLine, error) {
	q := `SELECT id, bom_id, line_no, component_product_id, qty_per::text, scrap_pct::text, op_no, alt_group
		FROM bom_lines WHERE bom_id=$1 ORDER BY line_no`
	rows, err := s.db(tx).Query(ctx, q, bomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.BOMLine
	for rows.Next() {
		var ln models.BOMLine
		if err := rows.Scan(&ln.ID, &ln.BomID, &ln.LineNo, &ln.ComponentProductID, &ln.QtyPer, &ln.ScrapPct, &ln.OpNo, &ln.AltGroup); err != nil {
			return nil, err
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}

// AddBOMLine добавляет строку.
func (s *Store) AddBOMLine(ctx context.Context, tx pgx.Tx, ln *models.BOMLine) error {
	q := `INSERT INTO bom_lines(id, bom_id, line_no, component_product_id, qty_per, scrap_pct, op_no, alt_group)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.db(tx).Exec(ctx, q, ln.ID, ln.BomID, ln.LineNo, ln.ComponentProductID, ln.QtyPer, ln.ScrapPct, ln.OpNo, ln.AltGroup)
	return err
}

// DeleteBOMLine удаляет строку.
func (s *Store) DeleteBOMLine(ctx context.Context, tx pgx.Tx, bomID, lineID uuid.UUID) error {
	q := `DELETE FROM bom_lines WHERE bom_id=$1 AND id=$2`
	tag, err := s.db(tx).Exec(ctx, q, bomID, lineID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteBOMLines удаляет все строки BOM.
func (s *Store) DeleteBOMLines(ctx context.Context, tx pgx.Tx, bomID uuid.UUID) error {
	_, err := s.db(tx).Exec(ctx, `DELETE FROM bom_lines WHERE bom_id=$1`, bomID)
	return err
}
