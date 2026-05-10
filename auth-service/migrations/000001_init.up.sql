CREATE TABLE tenants (
    code               VARCHAR(64)  PRIMARY KEY,
    name               VARCHAR(255) NOT NULL,
    keycloak_group_id  VARCHAR(64)  NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE user_cache (
    keycloak_id  VARCHAR(64)  PRIMARY KEY,
    tenant_code  VARCHAR(64)  NOT NULL REFERENCES tenants(code) ON DELETE CASCADE,
    username     VARCHAR(255) NOT NULL,
    email        VARCHAR(255),
    roles        JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_cache_tenant ON user_cache(tenant_code);
