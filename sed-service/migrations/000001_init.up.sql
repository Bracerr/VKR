CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE TABLE workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    order_no INT NOT NULL CHECK (order_no > 0),
    parallel_group INT,
    name VARCHAR(255) NOT NULL,
    required_role VARCHAR(64),
    required_user_sub VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (required_role IS NOT NULL OR required_user_sub IS NOT NULL)
);

CREATE INDEX idx_workflow_steps_wf ON workflow_steps(workflow_id, order_no);

CREATE TABLE document_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(512) NOT NULL,
    warehouse_action VARCHAR(16) NOT NULL DEFAULT 'NONE'
        CHECK (warehouse_action IN ('NONE', 'RESERVE', 'CONSUME', 'RECEIPT')),
    default_workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, code)
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(64) NOT NULL,
    type_id UUID NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
    number VARCHAR(64) NOT NULL,
    title VARCHAR(512) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'SIGNED', 'CANCELLED')),
    author_sub VARCHAR(255) NOT NULL,
    current_order_no INT,
    payload JSONB NOT NULL DEFAULT '{}',
    warehouse_ref JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, number)
);

CREATE INDEX idx_documents_tenant_status ON documents(tenant_code, status);
CREATE INDEX idx_documents_author ON documents(tenant_code, author_sub);

CREATE TABLE document_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    decision VARCHAR(16) NOT NULL DEFAULT 'PENDING'
        CHECK (decision IN ('PENDING', 'APPROVED', 'REJECTED')),
    decider_sub VARCHAR(255),
    comment TEXT,
    decided_at TIMESTAMPTZ,
    UNIQUE (document_id, step_id)
);

CREATE INDEX idx_doc_approvals_doc ON document_approvals(document_id);

CREATE TABLE document_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    object_key VARCHAR(1024) NOT NULL,
    original_name VARCHAR(512) NOT NULL,
    content_type VARCHAR(128),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    uploaded_by VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_doc_files_doc ON document_files(document_id);

CREATE TABLE document_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    actor_sub VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_doc_history_doc ON document_history(document_id, created_at);
