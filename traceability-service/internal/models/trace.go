package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TraceNode struct {
	ID         uuid.UUID       `json:"id"`
	TenantCode string          `json:"tenant_code"`
	NodeType   string          `json:"node_type"`
	ExternalID string          `json:"external_id"`
	Label      *string         `json:"label,omitempty"`
	Meta       json.RawMessage `json:"meta"`
	CreatedAt  time.Time       `json:"created_at"`
}

type TraceEdge struct {
	ID         uuid.UUID       `json:"id"`
	TenantCode string          `json:"tenant_code"`
	EdgeType   string          `json:"edge_type"`
	FromNodeID uuid.UUID       `json:"from_node_id"`
	ToNodeID   uuid.UUID       `json:"to_node_id"`
	Meta       json.RawMessage `json:"meta"`
	CreatedAt  time.Time       `json:"created_at"`
}

type TraceEvent struct {
	ID             uuid.UUID       `json:"id"`
	TenantCode     string          `json:"tenant_code"`
	EventType      string          `json:"event_type"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      time.Time       `json:"created_at"`
}

