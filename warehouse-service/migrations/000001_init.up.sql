-- Склад: справочники, движения, остатки, резервы, импорт

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Сентинел для товаров без партии (tracking NONE)
CREATE OR REPLACE FUNCTION wh_nil_batch() RETURNS uuid AS $$
  SELECT '00000000-0000-0000-0000-000000000000'::uuid;
$$ LANGUAGE sql IMMUTABLE;

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    sku VARCHAR(128) NOT NULL,
    name VARCHAR(512) NOT NULL,
    unit VARCHAR(32) NOT NULL DEFAULT 'pcs',
    tracking_mode VARCHAR(32) NOT NULL DEFAULT 'NONE'
        CHECK (tracking_mode IN ('NONE', 'BATCH', 'SERIAL', 'BATCH_AND_SERIAL')),
    has_expiration BOOLEAN NOT NULL DEFAULT false,
    valuation_method VARCHAR(16) NOT NULL DEFAULT 'AVERAGE'
        CHECK (valuation_method IN ('FIFO', 'AVERAGE', 'STANDARD')),
    default_currency CHAR(3) NOT NULL DEFAULT 'RUB',
    standard_cost NUMERIC(18,4),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, sku)
);

CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE TABLE bins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255),
    bin_type VARCHAR(32) NOT NULL DEFAULT 'STORAGE'
        CHECK (bin_type IN ('RECEIVING', 'STORAGE', 'PICKING', 'SHIPPING', 'QUARANTINE')),
    parent_bin_id UUID REFERENCES bins(id) ON DELETE SET NULL,
    capacity_qty NUMERIC(18,3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (warehouse_id, code)
);

CREATE INDEX idx_bins_tenant ON bins(tenant_code);

CREATE TABLE batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    series VARCHAR(128) NOT NULL,
    manufactured_at DATE,
    expires_at DATE,
    unit_cost NUMERIC(18,4),
    currency CHAR(3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, series)
);

CREATE INDEX idx_batches_expires ON batches(tenant_code, expires_at NULLS LAST);

CREATE TABLE serial_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_id UUID REFERENCES batches(id) ON DELETE SET NULL,
    serial_no VARCHAR(256) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'IN_STOCK'
        CHECK (status IN ('IN_STOCK', 'RESERVED', 'IN_TRANSIT', 'ISSUED', 'SCRAPPED')),
    warehouse_id UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    bin_id UUID REFERENCES bins(id) ON DELETE SET NULL,
    last_movement_id UUID,
    unit_cost NUMERIC(18,4),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, serial_no)
);

CREATE INDEX idx_serial_tenant ON serial_numbers(tenant_code);
CREATE INDEX idx_serial_wh ON serial_numbers(warehouse_id);

CREATE TABLE product_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price_type VARCHAR(16) NOT NULL
        CHECK (price_type IN ('SALE', 'PURCHASE', 'LIST', 'SPECIAL')),
    currency CHAR(3) NOT NULL,
    price NUMERIC(18,4) NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

CREATE INDEX idx_prices_product ON product_prices(tenant_code, product_id, price_type, valid_from);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    doc_type VARCHAR(32) NOT NULL
        CHECK (doc_type IN ('RECEIPT', 'ISSUE', 'TRANSFER', 'INVENTORY', 'RELOCATE', 'RESERVATION')),
    number VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED')),
    warehouse_from_id UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    warehouse_to_id UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    period_at DATE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, number)
);

CREATE TABLE document_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    batch_id UUID REFERENCES batches(id) ON DELETE SET NULL,
    serial_id UUID REFERENCES serial_numbers(id) ON DELETE SET NULL,
    qty NUMERIC(18,3) NOT NULL,
    unit_cost NUMERIC(18,4)
);

CREATE TABLE inventory_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    bin_id UUID NOT NULL REFERENCES bins(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL DEFAULT wh_nil_batch(),
    serial_id UUID REFERENCES serial_numbers(id) ON DELETE SET NULL,
    expected_qty NUMERIC(18,3) NOT NULL DEFAULT 0,
    counted_qty NUMERIC(18,3),
    diff NUMERIC(18,3) GENERATED ALWAYS AS (
        COALESCE(counted_qty, 0) - expected_qty
    ) STORED
);

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'RELEASED', 'CONSUMED', 'EXPIRED')),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    bin_id UUID REFERENCES bins(id) ON DELETE SET NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL DEFAULT wh_nil_batch(),
    serial_id UUID REFERENCES serial_numbers(id) ON DELETE SET NULL,
    qty NUMERIC(18,3) NOT NULL CHECK (qty > 0),
    reason VARCHAR(512),
    doc_ref VARCHAR(128),
    expires_at TIMESTAMPTZ,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_resv_status_exp ON reservations(status, expires_at);

CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    movement_type VARCHAR(32) NOT NULL
        CHECK (movement_type IN (
            'RECEIPT', 'ISSUE', 'TRANSFER_OUT', 'TRANSFER_IN',
            'RELOCATE_OUT', 'RELOCATE_IN', 'INVENTORY_ADJUST', 'RESERVE_CONSUMED'
        )),
    document_id UUID REFERENCES documents(id) ON DELETE SET NULL,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    bin_id UUID NOT NULL REFERENCES bins(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL DEFAULT wh_nil_batch(),
    serial_id UUID REFERENCES serial_numbers(id) ON DELETE SET NULL,
    qty NUMERIC(18,3) NOT NULL,
    unit_cost NUMERIC(18,4),
    value NUMERIC(18,4),
    currency CHAR(3),
    posted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    posted_by VARCHAR(255) NOT NULL
);

CREATE INDEX idx_mov_tenant_time ON stock_movements(tenant_code, posted_at);
CREATE INDEX idx_mov_product_time ON stock_movements(product_id, posted_at);
CREATE INDEX idx_mov_doc ON stock_movements(document_id);

CREATE TABLE stock_balances (
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    bin_id UUID NOT NULL REFERENCES bins(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL DEFAULT wh_nil_batch(),
    quantity NUMERIC(18,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved_qty NUMERIC(18,3) NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    value NUMERIC(18,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (warehouse_id, bin_id, product_id, batch_id),
    CHECK (reserved_qty <= quantity)
);

CREATE INDEX idx_bal_product ON stock_balances(product_id, batch_id);

CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    kind VARCHAR(32) NOT NULL
        CHECK (kind IN ('PRODUCTS', 'WAREHOUSES', 'BINS', 'OPENING_BALANCES', 'PRICES', 'SERIALS')),
    status VARCHAR(16) NOT NULL DEFAULT 'QUEUED'
        CHECK (status IN ('QUEUED', 'RUNNING', 'DONE', 'FAILED')),
    total INT NOT NULL DEFAULT 0,
    processed INT NOT NULL DEFAULT 0,
    errors JSONB,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE idempotency_keys (
    key VARCHAR(128) NOT NULL,
    tenant_code VARCHAR(64) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_code, key)
);
