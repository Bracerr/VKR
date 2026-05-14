CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    contacts JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE INDEX idx_customers_tenant ON customers(tenant_code);

CREATE TABLE sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    number VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','RELEASED','SHIPPED','CANCELLED')),
    customer_id UUID NOT NULL REFERENCES customers(id),
    created_by_sub VARCHAR(255) NOT NULL,
    ship_from_warehouse_id UUID,
    ship_from_bin_id UUID,
    note TEXT,
    sed_document_id UUID,
    reservations JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, number)
);

CREATE INDEX idx_so_tenant_status ON sales_orders(tenant_code, status);
CREATE INDEX idx_so_sed_doc ON sales_orders(tenant_code, sed_document_id);

CREATE TABLE sales_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    so_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    line_no INT NOT NULL CHECK (line_no > 0),
    product_id UUID NOT NULL,
    qty NUMERIC(24,8) NOT NULL CHECK (qty > 0),
    uom VARCHAR(16) NOT NULL DEFAULT 'pcs',
    reserved_qty NUMERIC(24,8) NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    shipped_qty NUMERIC(24,8) NOT NULL DEFAULT 0 CHECK (shipped_qty >= 0),
    note TEXT,
    UNIQUE (so_id, line_no)
);

CREATE INDEX idx_so_lines_so ON sales_order_lines(so_id);

CREATE TABLE shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    so_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE RESTRICT,
    warehouse_document_id UUID,
    status VARCHAR(16) NOT NULL DEFAULT 'POSTED' CHECK (status IN ('POSTED')),
    posted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, so_id, warehouse_document_id)
);

CREATE TABLE sales_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id UUID NOT NULL,
    actor_sub VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sales_hist_entity ON sales_history(tenant_code, entity_type, entity_id);

