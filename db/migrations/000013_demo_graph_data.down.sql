-- 000013_demo_graph_data.down.sql
-- Remove demo graph data

DELETE FROM edges
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
  AND from_doc_id IN (
    SELECT doc_id FROM documents
    WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
      AND stable_key LIKE 'demo/%'
  );

DELETE FROM documents
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
  AND stable_key LIKE 'demo/%';

-- Revert doc_type CHECK to original
ALTER TABLE documents
  DROP CONSTRAINT IF EXISTS documents_doc_type_check;

ALTER TABLE documents
  ADD CONSTRAINT documents_doc_type_check
  CHECK (doc_type IN (
    'report','adr','postmortem','snippet','glossary','guide',
    'policy','other'
  ));
