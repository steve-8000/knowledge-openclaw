-- 000008_bm25_index.down.sql
DROP INDEX IF EXISTS idx_chunks_bm25;
DROP INDEX IF EXISTS idx_chunks_seq;
ALTER TABLE chunks DROP COLUMN IF EXISTS chunk_seq;
