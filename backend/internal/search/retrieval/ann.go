package retrieval

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pgvector "github.com/pgvector/pgvector-go"
)

func SearchANN(ctx context.Context, db Querier, embedding []float32, tenantID uuid.UUID, topK int) ([]RankedChunk, error) {
	if topK <= 0 || len(embedding) == 0 {
		return []RankedChunk{}, nil
	}
	const sqlQuery = `
		SELECT
			c.chunk_id,
			c.doc_id,
			c.version_id,
			COALESCE(d.title, '') AS title,
			d.doc_type,
			d.status,
			d.tags,
			COALESCE(c.heading_path, '') AS heading_path,
			c.chunk_text,
			ROW_NUMBER() OVER (ORDER BY ce.embedding <=> $1::vector ASC) AS rank,
			(ce.embedding <=> $1::vector) AS score
		FROM chunk_embeddings ce
		JOIN chunks c
			ON c.tenant_id = ce.tenant_id
			AND c.chunk_id = ce.chunk_id
		JOIN documents d
			ON d.tenant_id = c.tenant_id
			AND d.doc_id = c.doc_id
		WHERE ce.tenant_id = $2
			AND d.status = 'published'
		ORDER BY score ASC
		LIMIT $3
	`

	vec := pgvector.NewVector(embedding)
	rows, err := db.Query(ctx, sqlQuery, vec, tenantID, topK)
	if err != nil {
		return nil, fmt.Errorf("query ann results: %w", err)
	}
	defer rows.Close()

	results := make([]RankedChunk, 0, topK)
	for rows.Next() {
		var r RankedChunk
		if err := rows.Scan(
			&r.ChunkID,
			&r.DocID,
			&r.VersionID,
			&r.Title,
			&r.DocType,
			&r.Status,
			&r.Tags,
			&r.HeadingPath,
			&r.ChunkText,
			&r.Rank,
			&r.Score,
		); err != nil {
			return nil, fmt.Errorf("scan ann row: %w", err)
		}

		rank := r.Rank
		r.ANNRank = &rank
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ann rows: %w", err)
	}

	return results, nil
}
