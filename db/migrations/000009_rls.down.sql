-- 000009_rls.down.sql
-- Remove RLS policies and disable RLS

DROP POLICY IF EXISTS tenant_isolation_documents ON documents;
DROP POLICY IF EXISTS tenant_isolation_doc_versions ON document_versions;
DROP POLICY IF EXISTS tenant_isolation_chunks ON chunks;
DROP POLICY IF EXISTS tenant_isolation_embeddings ON chunk_embeddings;
DROP POLICY IF EXISTS tenant_isolation_edges ON edges;
DROP POLICY IF EXISTS tenant_isolation_feedback ON search_feedback;
DROP POLICY IF EXISTS tenant_isolation_audit ON audit_log;
DROP POLICY IF EXISTS tenant_isolation_outbox ON outbox_events;

ALTER TABLE documents          DISABLE ROW LEVEL SECURITY;
ALTER TABLE document_versions  DISABLE ROW LEVEL SECURITY;
ALTER TABLE chunks             DISABLE ROW LEVEL SECURITY;
ALTER TABLE chunk_embeddings   DISABLE ROW LEVEL SECURITY;
ALTER TABLE edges              DISABLE ROW LEVEL SECURITY;
ALTER TABLE search_feedback    DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log          DISABLE ROW LEVEL SECURITY;
ALTER TABLE outbox_events      DISABLE ROW LEVEL SECURITY;
