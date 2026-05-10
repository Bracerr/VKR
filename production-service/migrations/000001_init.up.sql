CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE workcenters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    capacity_minutes_per_shift INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE INDEX idx_workcenters_tenant ON workcenters(tenant_code);

CREATE TABLE scrap_reasons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    UNIQUE (tenant_code, code)
);

CREATE TABLE boms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL,
    version INT NOT NULL CHECK (version > 0),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'SUBMITTED', 'APPROVED', 'ARCHIVED')),
    sed_document_id UUID,
    valid_from DATE,
    valid_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, product_id, version)
);

CREATE INDEX idx_boms_tenant_product ON boms(tenant_code, product_id);
CREATE INDEX idx_boms_sed_doc ON boms(tenant_code, sed_document_id);

CREATE TABLE bom_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bom_id UUID NOT NULL REFERENCES boms(id) ON DELETE CASCADE,
    line_no INT NOT NULL,
    component_product_id UUID NOT NULL,
    qty_per NUMERIC(24, 8) NOT NULL CHECK (qty_per > 0),
    scrap_pct NUMERIC(10, 4) NOT NULL DEFAULT 0 CHECK (scrap_pct >= 0),
    op_no INT NOT NULL CHECK (op_no > 0),
    alt_group VARCHAR(64),
    UNIQUE (bom_id, line_no)
);

CREATE TABLE routings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL,
    version INT NOT NULL CHECK (version > 0),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'SUBMITTED', 'APPROVED', 'ARCHIVED')),
    sed_document_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, product_id, version)
);

CREATE INDEX idx_routings_tenant_product ON routings(tenant_code, product_id);
CREATE INDEX idx_routings_sed_doc ON routings(tenant_code, sed_document_id);

CREATE TABLE routing_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    routing_id UUID NOT NULL REFERENCES routings(id) ON DELETE CASCADE,
    op_no INT NOT NULL CHECK (op_no > 0),
    workcenter_id UUID NOT NULL REFERENCES workcenters(id),
    name VARCHAR(512) NOT NULL,
    time_per_unit_min NUMERIC(12, 4),
    setup_time_min NUMERIC(12, 4),
    qc_required BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (routing_id, op_no)
);

CREATE TABLE production_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL,
    qty_planned NUMERIC(24, 8) NOT NULL CHECK (qty_planned > 0),
    qty_done NUMERIC(24, 8) NOT NULL DEFAULT 0,
    qty_scrap NUMERIC(24, 8) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'PLANNED'
        CHECK (status IN ('PLANNED', 'RELEASED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    bom_id UUID NOT NULL REFERENCES boms(id),
    routing_id UUID NOT NULL REFERENCES routings(id),
    warehouse_id UUID NOT NULL,
    default_bin_id UUID NOT NULL,
    reservations JSONB NOT NULL DEFAULT '[]',
    warehouse_receipt_doc_id UUID,
    start_plan TIMESTAMPTZ,
    finish_plan TIMESTAMPTZ,
    start_fact TIMESTAMPTZ,
    finish_fact TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE INDEX idx_prod_orders_tenant_status ON production_orders(tenant_code, status);

CREATE TABLE production_order_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    op_no INT NOT NULL CHECK (op_no > 0),
    workcenter_id UUID NOT NULL,
    name VARCHAR(512) NOT NULL,
    qty_planned NUMERIC(24, 8) NOT NULL,
    qty_good NUMERIC(24, 8) NOT NULL DEFAULT 0,
    qty_scrap NUMERIC(24, 8) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'STARTED', 'DONE')),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    UNIQUE (order_id, op_no)
);

CREATE INDEX idx_order_ops_order ON production_order_operations(order_id);

CREATE TABLE shift_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    order_operation_id UUID NOT NULL REFERENCES production_order_operations(id) ON DELETE CASCADE,
    shift_date DATE NOT NULL,
    shift_no SMALLINT NOT NULL CHECK (shift_no IN (1, 2, 3)),
    assignee_sub VARCHAR(255),
    qty_planned NUMERIC(24, 8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shift_tasks_tenant ON shift_tasks(tenant_code, shift_date);

CREATE TABLE production_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    order_operation_id UUID NOT NULL REFERENCES production_order_operations(id) ON DELETE CASCADE,
    reported_by_sub VARCHAR(255) NOT NULL,
    qty_good NUMERIC(24, 8) NOT NULL DEFAULT 0,
    qty_scrap NUMERIC(24, 8) NOT NULL DEFAULT 0,
    scrap_reason_code VARCHAR(64),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE production_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id UUID NOT NULL,
    actor_sub VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_prod_hist_entity ON production_history(tenant_code, entity_type, entity_id);
