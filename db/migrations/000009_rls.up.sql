-- 000009_rls.up.sql
-- Row Level Security policies for multi-tenant isolation
-- All queries MUST set: SET LOCAL app.tenant_id = '<uuid>';

-- Enable RLS on all tenant-scoped tables
ALTER TABLE documents          ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_versions  ENABLE ROW LEVEL SECURITY;
ALTER TABLE chunks             ENABLE ROW LEVEL SECURITY;
ALTER TABLE chunk_embeddings   ENABLE ROW LEVEL SECURITY;
ALTER TABLE edges              ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_feedback    ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log          ENABLE ROW LEVEL SECURITY;
ALTER TABLE outbox_events      ENABLE ROW LEVEL SECURITY;

-- Create application role (non-superuser) that RLS applies to
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'kidb_app') THEN
        CREATE ROLE kidb_app LOGIN PASSWORD 'kidb_app_secret';
    END IF;
END
$$;

-- Grant table permissions to app role
GRANT USAGE ON SCHEMA public TO kidb_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO kidb_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO kidb_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kidb_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO kidb_app;

-- RLS policies: tenant isolation via session variable
CREATE POLICY tenant_isolation_documents ON documents
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_doc_versions ON document_versions
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_chunks ON chunks
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_embeddings ON chunk_embeddings
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_edges ON edges
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_feedback ON search_feedback
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_audit ON audit_log
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY tenant_isolation_outbox ON outbox_events
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Superuser/migration role bypasses RLS (default Postgres behavior)
-- The kidb_app role is subject to RLS
ALTER TABLE documents          FORCE ROW LEVEL SECURITY;
ALTER TABLE document_versions  FORCE ROW LEVEL SECURITY;
ALTER TABLE chunks             FORCE ROW LEVEL SECURITY;
ALTER TABLE chunk_embeddings   FORCE ROW LEVEL SECURITY;
ALTER TABLE edges              FORCE ROW LEVEL SECURITY;
ALTER TABLE search_feedback    FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_log          FORCE ROW LEVEL SECURITY;
ALTER TABLE outbox_events      FORCE ROW LEVEL SECURITY;
