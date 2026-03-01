// Package outbox provides the transactional outbox pattern for event publishing.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/openclaw/ki-db/pkg/events"
)

// Writer writes events to the outbox table within an existing transaction.
type Writer struct{}

// NewWriter creates a new outbox writer.
func NewWriter() *Writer {
	return &Writer{}
}

// Write inserts an event into the outbox_events table within the given transaction.
// This ensures atomicity with the business data write.
func (w *Writer) Write(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, eventType events.EventType, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	envelope, err := events.NewEnvelope(eventType, tenantID, payload)
	if err != nil {
		return fmt.Errorf("create envelope: %w", err)
	}

	envelopeJSON, err := envelope.Marshal()
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (tenant_id, event_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`, tenantID, envelope.EventID, string(eventType), envelopeJSON)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	_ = data // payload already marshaled via envelope
	return nil
}
