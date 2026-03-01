-- 000012_rls_missing_ok.up.sql
-- Make RLS policies safe when app.tenant_id GUC is not yet set.
-- current_setting('app.tenant_id', true) returns NULL when missing,
-- so no rows match (safe deny-by-default).

DROP POLICY IF EXISTS tenant_isolation_documents ON documents;
CREATE POLICY tenant_isolation_documents ON documents
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_doc_versions ON document_versions;
CREATE POLICY tenant_isolation_doc_versions ON document_versions
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_chunks ON chunks;
CREATE POLICY tenant_isolation_chunks ON chunks
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_embeddings ON chunk_embeddings;
CREATE POLICY tenant_isolation_embeddings ON chunk_embeddings
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_edges ON edges;
CREATE POLICY tenant_isolation_edges ON edges
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_feedback ON search_feedback;
CREATE POLICY tenant_isolation_feedback ON search_feedback
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_audit ON audit_log;
CREATE POLICY tenant_isolation_audit ON audit_log
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

DROP POLICY IF EXISTS tenant_isolation_outbox ON outbox_events;
CREATE POLICY tenant_isolation_outbox ON outbox_events
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);
