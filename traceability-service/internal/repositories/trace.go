package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) InsertEvent(ctx context.Context, tx pgx.Tx, tenant, eventType string, idemKey *string, payload json.RawMessage) error {
	q := `INSERT INTO trace_events(tenant_code, event_type, idempotency_key, payload) VALUES ($1,$2,$3,$4)`
	_, err := s.db(tx).Exec(ctx, q, tenant, eventType, idemKey, payload)
	return err
}

func (s *Store) UpsertNode(ctx context.Context, tx pgx.Tx, tenant, nodeType, externalID string, label *string, meta json.RawMessage) (uuid.UUID, error) {
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	q := `
		INSERT INTO trace_nodes(tenant_code, node_type, external_id, label, meta)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_code, node_type, external_id) DO UPDATE SET
		  label = COALESCE(EXCLUDED.label, trace_nodes.label),
		  meta = trace_nodes.meta || EXCLUDED.meta
		RETURNING id`
	var id uuid.UUID
	err := s.db(tx).QueryRow(ctx, q, tenant, nodeType, externalID, label, meta).Scan(&id)
	return id, err
}

func (s *Store) UpsertEdge(ctx context.Context, tx pgx.Tx, tenant, edgeType string, fromID, toID uuid.UUID, meta json.RawMessage) error {
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	q := `
		INSERT INTO trace_edges(tenant_code, edge_type, from_node_id, to_node_id, meta)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_code, edge_type, from_node_id, to_node_id) DO UPDATE SET
		  meta = trace_edges.meta || EXCLUDED.meta`
	_, err := s.db(tx).Exec(ctx, q, tenant, edgeType, fromID, toID, meta)
	return err
}

func (s *Store) GetNodeID(ctx context.Context, tx pgx.Tx, tenant, nodeType, externalID string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := s.db(tx).QueryRow(ctx, `SELECT id FROM trace_nodes WHERE tenant_code=$1 AND node_type=$2 AND external_id=$3`, tenant, nodeType, externalID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &id, err
}

func (s *Store) GetNodeByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (map[string]any, error) {
	var nodeType, externalID string
	var label *string
	var meta []byte
	err := s.db(tx).QueryRow(ctx, `SELECT node_type, external_id, label, meta FROM trace_nodes WHERE tenant_code=$1 AND id=$2`, tenant, id).Scan(&nodeType, &externalID, &label, &meta)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m any
	_ = json.Unmarshal(meta, &m)
	return map[string]any{"id": id.String(), "node_type": nodeType, "external_id": externalID, "label": label, "meta": m}, nil
}

func (s *Store) ListEdgesTouching(ctx context.Context, tx pgx.Tx, tenant string, nodeIDs []uuid.UUID) ([]struct {
	ID         uuid.UUID
	EdgeType   string
	FromNodeID uuid.UUID
	ToNodeID   uuid.UUID
	Meta       []byte
}, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, edge_type, from_node_id, to_node_id, meta
		FROM trace_edges
		WHERE tenant_code=$1 AND (from_node_id = ANY($2) OR to_node_id = ANY($2))
	`, tenant, nodeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID         uuid.UUID
		EdgeType   string
		FromNodeID uuid.UUID
		ToNodeID   uuid.UUID
		Meta       []byte
	}
	for rows.Next() {
		var r struct {
			ID         uuid.UUID
			EdgeType   string
			FromNodeID uuid.UUID
			ToNodeID   uuid.UUID
			Meta       []byte
		}
		if err := rows.Scan(&r.ID, &r.EdgeType, &r.FromNodeID, &r.ToNodeID, &r.Meta); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

