-- 000008_bm25_index.up.sql
-- BM25 index using pg_search (ParadeDB) for high-quality keyword search
-- Note: Only one BM25 index per table allowed. key_field must be unique.

-- We need a unique single-column key for pg_search BM25 index.
-- chunk_id alone is unique within a tenant, but pg_search needs a global unique key.
-- We use chunk_id (uuid PK within tenant) as the key field.

-- First, add a surrogate bigserial for pg_search key_field requirement
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS chunk_seq bigserial;
CREATE UNIQUE INDEX IF NOT EXISTS idx_chunks_seq ON chunks (chunk_seq);

-- BM25 index on chunks for full-text search
CREATE INDEX idx_chunks_bm25
    ON chunks
    USING bm25 (
        chunk_seq,
        chunk_text,
        heading_path
    )
    WITH (key_field = 'chunk_seq');
