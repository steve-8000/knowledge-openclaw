-- 000004_embeddings.up.sql
-- Vector embeddings for chunks with HNSW index (1024 dims, cosine distance)

CREATE TABLE chunk_embeddings (
    tenant_id       uuid    NOT NULL,
    chunk_id        uuid    NOT NULL,
    embedding_model text    NOT NULL,
    dims            int     NOT NULL DEFAULT 1024,
    embedding       vector(1024),
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, chunk_id, embedding_model),
    FOREIGN KEY (tenant_id, chunk_id) REFERENCES chunks(tenant_id, chunk_id) ON DELETE CASCADE
);

-- HNSW index for approximate nearest neighbor search (cosine distance)
CREATE INDEX idx_chunk_embeddings_hnsw
    ON chunk_embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 200);

-- Tenant-filtered partial index for published documents only
-- (will be useful when combined with status filter queries)
