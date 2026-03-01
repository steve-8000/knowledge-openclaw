package retrieval

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func SearchBM25(ctx context.Context, db Querier, query string, tenantID uuid.UUID, topK int) ([]RankedChunk, error) {
	if topK <= 0 {
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
			ROW_NUMBER() OVER (ORDER BY pdb.score(c.chunk_seq) DESC) AS rank,
			pdb.score(c.chunk_seq) AS score
		FROM chunks c
		JOIN documents d
			ON d.tenant_id = c.tenant_id
			AND d.doc_id = c.doc_id
		WHERE c.tenant_id = $2
			AND d.status = 'published'
			AND c.chunk_text ||| $1
		ORDER BY score DESC
		LIMIT $3
	`

	rows, err := db.Query(ctx, sqlQuery, query, tenantID, topK)
	if err != nil {
		return nil, fmt.Errorf("query bm25 results: %w", err)
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
			return nil, fmt.Errorf("scan bm25 row: %w", err)
		}

		rank := r.Rank
		r.BM25Rank = &rank
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bm25 rows: %w", err)
	}

	return results, nil
}
