// Package idempotency provides event deduplication for workers.
package idempotency

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Store manages idempotency keys in the database.
type Store struct{}

// NewStore creates a new idempotency store.
func NewStore() *Store {
	return &Store{}
}

// CheckAndMark returns true if the event has already been processed.
// If not, it marks it as processed within the given transaction.
func (s *Store) CheckAndMark(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, workerName string) (bool, error) {
	// Try to insert; if already exists, it's a duplicate
	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_keys (event_id, worker_name, processed_at)
		VALUES ($1, $2, now())
		ON CONFLICT (event_id, worker_name) DO NOTHING
	`, eventID, workerName)
	if err != nil {
		return false, fmt.Errorf("check idempotency: %w", err)
	}

	// If no rows affected, it was already processed
	return tag.RowsAffected() == 0, nil
}
