-- 000003_chunks.up.sql
-- Chunks: search-unit text segments from document versions

CREATE TABLE chunks (
    tenant_id    uuid        NOT NULL,
    chunk_id     uuid        NOT NULL DEFAULT uuid_generate_v4(),
    doc_id       uuid        NOT NULL,
    version_id   uuid        NOT NULL,
    ordinal      int         NOT NULL,
    heading_path text,
    chunk_text   text        NOT NULL,
    token_count  int,
    chunk_sha256 bytea       NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, chunk_id),
    FOREIGN KEY (tenant_id, doc_id, version_id)
        REFERENCES document_versions(tenant_id, doc_id, version_id) ON DELETE CASCADE
);

-- Ordering index for chunk retrieval
CREATE INDEX idx_chunks_doc_version ON chunks (tenant_id, doc_id, version_id, ordinal);

-- Native FTS fallback (used if pg_search unavailable)
ALTER TABLE chunks ADD COLUMN chunk_tsv tsvector
    GENERATED ALWAYS AS (to_tsvector('english', chunk_text)) STORED;
CREATE INDEX idx_chunks_tsv ON chunks USING gin (chunk_tsv);
