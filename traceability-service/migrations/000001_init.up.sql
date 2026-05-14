CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE trace_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128),
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, idempotency_key)
);

CREATE INDEX idx_trace_events_tenant_time ON trace_events(tenant_code, created_at);

CREATE TABLE trace_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    node_type VARCHAR(32) NOT NULL,
    external_id VARCHAR(128) NOT NULL,
    label VARCHAR(512),
    meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, node_type, external_id)
);

CREATE INDEX idx_trace_nodes_tenant_type ON trace_nodes(tenant_code, node_type);
CREATE INDEX idx_trace_nodes_lookup ON trace_nodes(tenant_code, node_type, external_id);

CREATE TABLE trace_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    edge_type VARCHAR(64) NOT NULL,
    from_node_id UUID NOT NULL REFERENCES trace_nodes(id) ON DELETE CASCADE,
    to_node_id UUID NOT NULL REFERENCES trace_nodes(id) ON DELETE CASCADE,
    meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, edge_type, from_node_id, to_node_id)
);

CREATE INDEX idx_trace_edges_from ON trace_edges(tenant_code, from_node_id);
CREATE INDEX idx_trace_edges_to ON trace_edges(tenant_code, to_node_id);

