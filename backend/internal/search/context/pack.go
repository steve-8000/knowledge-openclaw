package context

import (
	"github.com/openclaw/ki-db/internal/search/retrieval"
	"github.com/openclaw/ki-db/pkg/models"
)

func BuildContextPack(query string, results []retrieval.RankedChunk, limit int) *models.ContextPack {
	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}

	trimmed := results[:limit]
	packed := &models.ContextPack{
		Query:     query,
		Results:   make([]models.SearchResult, 0, len(trimmed)),
		Citations: make([]models.Citation, 0, len(trimmed)),
	}

	seenDocs := make(map[string]struct{}, len(trimmed))
	for _, chunk := range trimmed {
		packed.Results = append(packed.Results, models.SearchResult{
			ChunkID:     chunk.ChunkID,
			DocID:       chunk.DocID,
			Title:       chunk.Title,
			HeadingPath: chunk.HeadingPath,
			ChunkText:   chunk.ChunkText,
			BM25Rank:    chunk.BM25Rank,
			ANNRank:     chunk.ANNRank,
			RRFScore:    chunk.Score,
			RerankScore: chunk.RerankScore,
		})

		docKey := chunk.DocID.String()
		if _, ok := seenDocs[docKey]; ok {
			continue
		}
		seenDocs[docKey] = struct{}{}

		packed.Citations = append(packed.Citations, models.Citation{
			DocID:       chunk.DocID,
			VersionID:   chunk.VersionID,
			Title:       chunk.Title,
			HeadingPath: chunk.HeadingPath,
			ChunkID:     chunk.ChunkID,
			Relevance:   relevanceType(chunk),
		})
	}

	return packed
}

func relevanceType(chunk retrieval.RankedChunk) string {
	switch {
	case chunk.BM25Rank != nil && chunk.ANNRank != nil:
		return "both"
	case chunk.BM25Rank != nil:
		return "keyword"
	case chunk.ANNRank != nil:
		return "semantic"
	default:
		return "both"
	}
}
