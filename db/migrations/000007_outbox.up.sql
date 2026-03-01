-- 000007_outbox.up.sql
-- Transactional outbox for event-driven indexing pipeline

CREATE TABLE outbox_events (
    tenant_id    uuid        NOT NULL,
    event_id     uuid        PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type   text        NOT NULL CHECK (event_type IN (
        'DocumentUpserted','VersionCreated','DocumentParsed',
        'ChunksCreated','EmbeddingsGenerated','GraphUpdated',
        'IndexJobFailed','FeedbackReceived'
    )),
    payload      jsonb       NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz
);

-- Index for polling unpublished events (outbox relay)
CREATE INDEX idx_outbox_unpublished
    ON outbox_events (created_at ASC)
    WHERE published_at IS NULL;

-- Index for monitoring outbox lag
CREATE INDEX idx_outbox_tenant_created
    ON outbox_events (tenant_id, created_at DESC);
