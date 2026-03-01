-- 000006_feedback_audit.up.sql
-- Search feedback and audit log tables

CREATE TABLE search_feedback (
    tenant_id         uuid        NOT NULL,
    feedback_id       uuid        PRIMARY KEY DEFAULT uuid_generate_v4(),
    query             text        NOT NULL,
    selected_chunk_id uuid,
    helpful           boolean,
    note              text,
    created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_search_feedback_tenant ON search_feedback (tenant_id, created_at DESC);

CREATE TABLE audit_log (
    tenant_id  uuid        NOT NULL,
    audit_id   uuid        PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor      text        NOT NULL,
    action     text        NOT NULL CHECK (action IN (
        'tag_update','status_change','edge_create','edge_delete',
        'merge','doc_create','doc_update','doc_delete',
        'version_create','supersede','archive'
    )),
    payload    jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_tenant ON audit_log (tenant_id, created_at DESC);
CREATE INDEX idx_audit_log_action ON audit_log (tenant_id, action, created_at DESC);
