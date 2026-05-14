package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DocumentLineRow struct {
	ProductID uuid.UUID
	BatchID   uuid.UUID
	SerialID  *uuid.UUID
	Qty       string
}

// ListDocumentLinesByDocument возвращает строки документа (с проверкой tenant через documents).
func (s *Store) ListDocumentLinesByDocument(ctx context.Context, tx pgx.Tx, tenant string, docID uuid.UUID) ([]DocumentLineRow, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT dl.product_id, dl.batch_id, dl.serial_id, dl.qty::text
		FROM document_lines dl
		JOIN documents d ON d.id = dl.document_id
		WHERE dl.document_id=$1 AND d.tenant_code=$2
		ORDER BY dl.id
	`, docID, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DocumentLineRow
	for rows.Next() {
		var r DocumentLineRow
		if err := rows.Scan(&r.ProductID, &r.BatchID, &r.SerialID, &r.Qty); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

