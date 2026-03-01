package retrieval

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Querier is the common interface for pgx.Tx and pgxpool.Pool.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type RankedChunk struct {
	ChunkID     uuid.UUID
	DocID       uuid.UUID
	VersionID   uuid.UUID
	Title       string
	DocType     string
	Status      string
	Tags        []string
	HeadingPath string
	ChunkText   string
	Rank        int
	Score       float64
	BM25Rank    *int
	ANNRank     *int
	RerankScore *float64
}
