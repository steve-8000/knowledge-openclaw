// Package events defines the canonical event types for the ki-db indexing pipeline.
// All events are serialized as JSON and published via NATS JetStream.
// Subject hierarchy: idx.doc.{event_type}
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType enumerates all pipeline event types.
type EventType string

const (
	TypeDocumentUpserted    EventType = "DocumentUpserted"
	TypeVersionCreated      EventType = "VersionCreated"
	TypeDocumentParsed      EventType = "DocumentParsed"
	TypeChunksCreated       EventType = "ChunksCreated"
	TypeEmbeddingsGenerated EventType = "EmbeddingsGenerated"
	TypeGraphUpdated        EventType = "GraphUpdated"
	TypeIndexJobFailed      EventType = "IndexJobFailed"
	TypeFeedbackReceived    EventType = "FeedbackReceived"
)

// Subject returns the NATS subject for this event type.
func (t EventType) Subject() string {
	return "idx.doc." + string(t)
}

// Envelope wraps every event with routing and tracing metadata.
type Envelope struct {
	EventID   uuid.UUID       `json:"event_id"`
	EventType EventType       `json:"event_type"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// NewEnvelope creates a new event envelope.
func NewEnvelope(eventType EventType, tenantID uuid.UUID, payload any) (*Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		EventID:   uuid.New(),
		EventType: eventType,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Payload:   data,
	}, nil
}

// Marshal serializes the envelope to JSON.
func (e *Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEnvelope deserializes a JSON envelope.
func UnmarshalEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// --- Payload types ---

// DocumentUpsertedPayload is emitted when a document is created or updated.
type DocumentUpsertedPayload struct {
	DocID     uuid.UUID `json:"doc_id"`
	StableKey string    `json:"stable_key"`
	Title     string    `json:"title,omitempty"`
	DocType   string    `json:"doc_type"`
	Status    string    `json:"status"`
}

// VersionCreatedPayload is emitted when a new version of a document is created.
type VersionCreatedPayload struct {
	DocID         uuid.UUID `json:"doc_id"`
	VersionID     uuid.UUID `json:"version_id"`
	VersionNo     int64     `json:"version_no"`
	ContentURI    string    `json:"content_uri,omitempty"`
	ContentSHA256 string    `json:"content_sha256"`
}

// DocumentParsedPayload is emitted after parsing and normalizing a document version.
type DocumentParsedPayload struct {
	DocID          uuid.UUID      `json:"doc_id"`
	VersionID      uuid.UUID      `json:"version_id"`
	NormalizedText string         `json:"normalized_text,omitempty"` // may be omitted if stored in DB
	Links          []string       `json:"links,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// ChunksCreatedPayload is emitted after chunking a parsed document.
type ChunksCreatedPayload struct {
	DocID     uuid.UUID   `json:"doc_id"`
	VersionID uuid.UUID   `json:"version_id"`
	ChunkIDs  []uuid.UUID `json:"chunk_ids"`
	Count     int         `json:"count"`
}

// EmbeddingsGeneratedPayload is emitted after embedding chunks.
type EmbeddingsGeneratedPayload struct {
	DocID          uuid.UUID   `json:"doc_id"`
	VersionID      uuid.UUID   `json:"version_id"`
	ChunkIDs       []uuid.UUID `json:"chunk_ids"`
	EmbeddingModel string      `json:"embedding_model"`
	Dims           int         `json:"dims"`
}

// GraphUpdatedPayload is emitted after updating knowledge graph edges.
type GraphUpdatedPayload struct {
	DocID        uuid.UUID `json:"doc_id"`
	EdgesCreated int       `json:"edges_created"`
	EdgesRemoved int       `json:"edges_removed"`
	Relations    []string  `json:"relations,omitempty"` // e.g., ["supersedes", "references"]
}

// IndexJobFailedPayload is emitted when an indexing step fails after retries.
type IndexJobFailedPayload struct {
	DocID     uuid.UUID `json:"doc_id"`
	VersionID uuid.UUID `json:"version_id,omitempty"`
	Stage     string    `json:"stage"` // parser|chunker|embedder|graph|quality
	Error     string    `json:"error"`
	Attempts  int       `json:"attempts"`
}

// FeedbackReceivedPayload is emitted when search feedback is submitted.
type FeedbackReceivedPayload struct {
	Query           string    `json:"query"`
	SelectedChunkID uuid.UUID `json:"selected_chunk_id,omitempty"`
	Helpful         *bool     `json:"helpful,omitempty"`
	Note            string    `json:"note,omitempty"`
}
