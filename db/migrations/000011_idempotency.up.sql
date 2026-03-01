-- 000011_idempotency.up.sql
-- Idempotency keys for worker event deduplication

CREATE TABLE idempotency_keys (
    event_id     uuid NOT NULL,
    worker_name  text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, worker_name)
);

-- Auto-cleanup old keys (older than 7 days)
CREATE INDEX idx_idempotency_cleanup ON idempotency_keys (processed_at);
