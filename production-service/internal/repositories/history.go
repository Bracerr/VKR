package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InsertHistory аудит.
func (s *Store) InsertHistory(ctx context.Context, tx pgx.Tx, tenant string, entityType string, entityID uuid.UUID, actorSub, action string, payload json.RawMessage) error {
	q := `INSERT INTO production_history(id, tenant_code, entity_type, entity_id, actor_sub, action, payload)
		VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$6)`
	_, err := s.db(tx).Exec(ctx, q, tenant, entityType, entityID, actorSub, action, payload)
	return err
}
