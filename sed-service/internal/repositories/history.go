package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/sed-service/internal/models"
)

// ListDocumentHistory история.
func (s *Store) ListDocumentHistory(ctx context.Context, tx pgx.Tx, docID uuid.UUID) ([]models.DocumentHistory, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, document_id, actor_sub, action, payload, created_at
		FROM document_history WHERE document_id = $1 ORDER BY created_at
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DocumentHistory
	for rows.Next() {
		var h models.DocumentHistory
		if err := rows.Scan(&h.ID, &h.DocumentID, &h.ActorSub, &h.Action, &h.Payload, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
