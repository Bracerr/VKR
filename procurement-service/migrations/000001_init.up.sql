CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    inn VARCHAR(32),
    kpp VARCHAR(32),
    contacts JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE INDEX idx_suppliers_tenant ON suppliers(tenant_code);

CREATE TABLE purchase_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    number VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED','CANCELLED')),
    created_by_sub VARCHAR(255) NOT NULL,
    needed_by DATE,
    note TEXT,
    sed_document_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, number)
);

CREATE INDEX idx_pr_tenant_status ON purchase_requests(tenant_code, status);
CREATE INDEX idx_pr_sed_doc ON purchase_requests(tenant_code, sed_document_id);

CREATE TABLE purchase_request_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_id UUID NOT NULL REFERENCES purchase_requests(id) ON DELETE CASCADE,
    line_no INT NOT NULL CHECK (line_no > 0),
    product_id UUID NOT NULL,
    qty NUMERIC(24,8) NOT NULL CHECK (qty > 0),
    uom VARCHAR(16) NOT NULL DEFAULT 'pcs',
    target_warehouse_id UUID,
    target_bin_id UUID,
    note TEXT,
    UNIQUE (pr_id, line_no)
);

CREATE INDEX idx_pr_lines_pr ON purchase_request_lines(pr_id);

CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    number VARCHAR(64) NOT NULL,
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED','RELEASED','PARTIALLY_RECEIVED','RECEIVED','CANCELLED')),
    created_by_sub VARCHAR(255) NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'RUB',
    expected_at DATE,
    sed_document_id UUID,
    source_pr_id UUID REFERENCES purchase_requests(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, number)
);

CREATE INDEX idx_po_tenant_status ON purchase_orders(tenant_code, status);
CREATE INDEX idx_po_sed_doc ON purchase_orders(tenant_code, sed_document_id);

CREATE TABLE purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    line_no INT NOT NULL CHECK (line_no > 0),
    product_id UUID NOT NULL,
    qty_ordered NUMERIC(24,8) NOT NULL CHECK (qty_ordered > 0),
    qty_received NUMERIC(24,8) NOT NULL DEFAULT 0 CHECK (qty_received >= 0),
    price NUMERIC(24,8) NOT NULL DEFAULT 0 CHECK (price >= 0),
    vat_rate NUMERIC(10,4) NOT NULL DEFAULT 0 CHECK (vat_rate >= 0),
    target_warehouse_id UUID,
    target_bin_id UUID,
    UNIQUE (po_id, line_no)
);

CREATE INDEX idx_po_lines_po ON purchase_order_lines(po_id);

CREATE TABLE receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE RESTRICT,
    warehouse_document_id UUID NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'POSTED' CHECK (status IN ('POSTED')),
    posted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, po_id, warehouse_document_id)
);

CREATE TABLE procurement_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id UUID NOT NULL,
    actor_sub VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_proc_hist_entity ON procurement_history(tenant_code, entity_type, entity_id);

