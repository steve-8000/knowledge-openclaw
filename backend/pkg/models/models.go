// Package models defines the core domain entities for ki-db.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant system.
type Tenant struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Name     string    `json:"name"`
}

// Document represents a knowledge document.
type Document struct {
	TenantID   uuid.UUID      `json:"tenant_id"`
	DocID      uuid.UUID      `json:"doc_id"`
	StableKey  string         `json:"stable_key"`
	Title      string         `json:"title,omitempty"`
	DocType    string         `json:"doc_type"`   // report|adr|postmortem|snippet|glossary
	Status     string         `json:"status"`     // inbox|published|deprecated|archived
	Confidence string         `json:"confidence"` // high|med|low
	Owners     []string       `json:"owners"`
	Tags       []string       `json:"tags"`
	Source     map[string]any `json:"source"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// DocumentVersion represents a specific version of a document.
type DocumentVersion struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	DocID          uuid.UUID `json:"doc_id"`
	VersionID      uuid.UUID `json:"version_id"`
	VersionNo      int64     `json:"version_no"`
	ContentURI     string    `json:"content_uri,omitempty"`
	RawText        string    `json:"raw_text,omitempty"`
	NormalizedText string    `json:"normalized_text,omitempty"`
	ContentSHA256  []byte    `json:"content_sha256"`
	CreatedAt      time.Time `json:"created_at"`
}

// Chunk represents a search-unit chunk of a document version.
type Chunk struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	ChunkID     uuid.UUID `json:"chunk_id"`
	DocID       uuid.UUID `json:"doc_id"`
	VersionID   uuid.UUID `json:"version_id"`
	Ordinal     int       `json:"ordinal"`
	HeadingPath string    `json:"heading_path,omitempty"`
	ChunkText   string    `json:"chunk_text"`
	TokenCount  int       `json:"token_count,omitempty"`
	ChunkSHA256 []byte    `json:"chunk_sha256"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChunkEmbedding represents a vector embedding for a chunk.
type ChunkEmbedding struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	ChunkID        uuid.UUID `json:"chunk_id"`
	EmbeddingModel string    `json:"embedding_model"`
	Dims           int       `json:"dims"`
	Embedding      []float32 `json:"embedding"`
	CreatedAt      time.Time `json:"created_at"`
}

// Edge represents a knowledge graph edge between documents.
type Edge struct {
	TenantID      uuid.UUID      `json:"tenant_id"`
	EdgeID        uuid.UUID      `json:"edge_id"`
	FromDocID     uuid.UUID      `json:"from_doc_id"`
	ToDocID       *uuid.UUID     `json:"to_doc_id,omitempty"`
	ToExternalKey string         `json:"to_external_key,omitempty"`
	Relation      string         `json:"relation"` // links_to|references|supersedes|duplicates|contradicts
	Evidence      map[string]any `json:"evidence,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// SearchFeedback represents user feedback on search results.
type SearchFeedback struct {
	TenantID        uuid.UUID  `json:"tenant_id"`
	FeedbackID      uuid.UUID  `json:"feedback_id"`
	Query           string     `json:"query"`
	SelectedChunkID *uuid.UUID `json:"selected_chunk_id,omitempty"`
	Helpful         *bool      `json:"helpful,omitempty"`
	Note            string     `json:"note,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// AuditLog represents an audit trail entry.
type AuditLog struct {
	TenantID  uuid.UUID      `json:"tenant_id"`
	AuditID   uuid.UUID      `json:"audit_id"`
	Actor     string         `json:"actor"`
	Action    string         `json:"action"` // tag_update|status_change|edge_create|merge
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}

// OutboxEvent represents a transactional outbox event.
type OutboxEvent struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	EventID     uuid.UUID  `json:"event_id"`
	EventType   string     `json:"event_type"`
	Payload     []byte     `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// --- Search result types ---

// SearchResult represents a single search result with scoring info.
type SearchResult struct {
	ChunkID     uuid.UUID `json:"chunk_id"`
	DocID       uuid.UUID `json:"doc_id"`
	Title       string    `json:"title,omitempty"`
	HeadingPath string    `json:"heading_path,omitempty"`
	ChunkText   string    `json:"chunk_text"`
	BM25Rank    *int      `json:"bm25_rank,omitempty"`
	ANNRank     *int      `json:"ann_rank,omitempty"`
	RRFScore    float64   `json:"rrf_score"`
	RerankScore *float64  `json:"rerank_score,omitempty"`
}

// ContextPack represents a packed context for RAG with citations.
type ContextPack struct {
	Query     string         `json:"query"`
	Results   []SearchResult `json:"results"`
	Citations []Citation     `json:"citations"`
}

// Citation represents a source citation for a search result.
type Citation struct {
	DocID       uuid.UUID `json:"doc_id"`
	VersionID   uuid.UUID `json:"version_id"`
	Title       string    `json:"title"`
	HeadingPath string    `json:"heading_path,omitempty"`
	ChunkID     uuid.UUID `json:"chunk_id"`
	Relevance   string    `json:"relevance"` // keyword|semantic|both
}

// --- Document status constants ---

const (
	StatusInbox      = "inbox"
	StatusPublished  = "published"
	StatusDeprecated = "deprecated"
	StatusArchived   = "archived"
)

// --- Confidence level constants ---

const (
	ConfidenceHigh = "high"
	ConfidenceMed  = "med"
	ConfidenceLow  = "low"
)

// --- Relation type constants ---

const (
	RelationLinksTo     = "links_to"
	RelationReferences  = "references"
	RelationSupersedes  = "supersedes"
	RelationDuplicates  = "duplicates"
	RelationContradicts = "contradicts"
)
