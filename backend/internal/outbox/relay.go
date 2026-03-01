package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/pkg/events"
)

// Relay polls the outbox table and publishes unpublished events to NATS JetStream.
type Relay struct {
	pool         *pgxpool.Pool
	nats         *natsclient.Client
	batchSize    int
	pollInterval time.Duration
	logger       *slog.Logger
}

// NewRelay creates a new outbox relay.
func NewRelay(pool *pgxpool.Pool, nats *natsclient.Client, batchSize int, pollInterval time.Duration, logger *slog.Logger) *Relay {
	return &Relay{
		pool:         pool,
		nats:         nats,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		logger:       logger,
	}
}

// Run starts the relay loop. It blocks until the context is cancelled.
func (r *Relay) Run(ctx context.Context) error {
	r.logger.Info("outbox relay started", "batch_size", r.batchSize, "poll_interval", r.pollInterval)

	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("outbox relay stopping")
			return ctx.Err()
		case <-ticker.C:
			published, err := r.pollAndPublish(ctx)
			if err != nil {
				r.logger.Error("outbox relay poll error", "error", err)
				continue
			}
			if published > 0 {
				r.logger.Info("outbox relay published events", "count", published)
			}
		}
	}
}

func (r *Relay) pollAndPublish(ctx context.Context) (int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, event_type, payload
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, r.batchSize)
	if err != nil {
		return 0, fmt.Errorf("query outbox: %w", err)
	}
	defer rows.Close()

	published := 0
	for rows.Next() {
		var eventID string
		var eventType string
		var payload []byte

		if err := rows.Scan(&eventID, &eventType, &payload); err != nil {
			return published, fmt.Errorf("scan outbox row: %w", err)
		}

		// Publish to NATS with MsgID for dedup
		subject := events.EventType(eventType).Subject()
		if err := r.nats.Publish(ctx, subject, payload, eventID); err != nil {
			r.logger.Error("failed to publish event",
				"event_id", eventID,
				"event_type", eventType,
				"error", err,
			)
			continue
		}

		// Mark as published
		_, err := r.pool.Exec(ctx, `
			UPDATE outbox_events SET published_at = now() WHERE event_id = $1
		`, eventID)
		if err != nil {
			r.logger.Error("failed to mark event published",
				"event_id", eventID,
				"error", err,
			)
			continue
		}

		published++
	}

	return published, rows.Err()
}
