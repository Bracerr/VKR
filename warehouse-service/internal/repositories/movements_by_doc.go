package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MovementByDocRow struct {
	ProductID uuid.UUID
	BatchID   uuid.UUID
	SerialID  *uuid.UUID
	Qty       string
	PostedAt  time.Time
}

func (s *Store) ListMovementsByDocument(ctx context.Context, tx pgx.Tx, tenant string, docID uuid.UUID) ([]MovementByDocRow, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT product_id, batch_id, serial_id, qty::text, posted_at
		FROM stock_movements
		WHERE tenant_code=$1 AND document_id=$2
		ORDER BY posted_at, id
	`, tenant, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MovementByDocRow
	for rows.Next() {
		var r MovementByDocRow
		if err := rows.Scan(&r.ProductID, &r.BatchID, &r.SerialID, &r.Qty, &r.PostedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

