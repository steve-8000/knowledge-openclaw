-- 000002_core_tables.up.sql
-- Core tables: tenants, documents, document_versions

CREATE TABLE tenants (
    tenant_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name      text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE documents (
    tenant_id   uuid        NOT NULL REFERENCES tenants(tenant_id),
    doc_id      uuid        NOT NULL DEFAULT uuid_generate_v4(),
    stable_key  text        NOT NULL,
    title       text,
    doc_type    text        NOT NULL CHECK (doc_type IN ('report','adr','postmortem','snippet','glossary','guide','policy','other')),
    status      text        NOT NULL DEFAULT 'inbox' CHECK (status IN ('inbox','published','deprecated','archived')),
    confidence  text        NOT NULL DEFAULT 'med' CHECK (confidence IN ('high','med','low')),
    owners      jsonb       NOT NULL DEFAULT '[]',
    tags        jsonb       NOT NULL DEFAULT '[]',
    source      jsonb       NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, doc_id),
    UNIQUE (tenant_id, stable_key)
);

CREATE TABLE document_versions (
    tenant_id      uuid        NOT NULL,
    doc_id         uuid        NOT NULL,
    version_id     uuid        NOT NULL DEFAULT uuid_generate_v4(),
    version_no     bigint      NOT NULL,
    content_uri    text,
    raw_text       text,
    normalized_text text,
    content_sha256 bytea       NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, doc_id, version_id),
    UNIQUE (tenant_id, doc_id, version_no),
    FOREIGN KEY (tenant_id, doc_id) REFERENCES documents(tenant_id, doc_id) ON DELETE CASCADE
);

-- Index for fast version lookups
CREATE INDEX idx_doc_versions_doc ON document_versions (tenant_id, doc_id, version_no DESC);
