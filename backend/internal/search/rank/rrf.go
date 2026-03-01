package rank

import (
	"sort"

	"github.com/openclaw/ki-db/internal/search/retrieval"
)

func FuseRRF(bm25Results, annResults []retrieval.RankedChunk, k int) []retrieval.RankedChunk {
	if k <= 0 {
		k = 60
	}

	byChunk := make(map[string]retrieval.RankedChunk, len(bm25Results)+len(annResults))

	for i, chunk := range bm25Results {
		rank := chunk.Rank
		if rank <= 0 {
			rank = i + 1
		}

		key := chunk.ChunkID.String()
		agg, ok := byChunk[key]
		if ok {
			agg.Score += 1.0 / float64(k+rank)
			if agg.BM25Rank == nil {
				r := rank
				agg.BM25Rank = &r
			}
			byChunk[key] = agg
			continue
		}

		r := rank
		chunk.BM25Rank = &r
		chunk.Score = 1.0 / float64(k+rank)
		chunk.ANNRank = nil
		chunk.RerankScore = nil
		byChunk[key] = chunk
	}

	for i, chunk := range annResults {
		rank := chunk.Rank
		if rank <= 0 {
			rank = i + 1
		}

		key := chunk.ChunkID.String()
		agg, ok := byChunk[key]
		if ok {
			agg.Score += 1.0 / float64(k+rank)
			if agg.ANNRank == nil {
				r := rank
				agg.ANNRank = &r
			}
			byChunk[key] = agg
			continue
		}

		r := rank
		chunk.ANNRank = &r
		chunk.BM25Rank = nil
		chunk.Score = 1.0 / float64(k+rank)
		chunk.RerankScore = nil
		byChunk[key] = chunk
	}

	fused := make([]retrieval.RankedChunk, 0, len(byChunk))
	for _, chunk := range byChunk {
		fused = append(fused, chunk)
	}

	sort.SliceStable(fused, func(i, j int) bool {
		if fused[i].Score == fused[j].Score {
			return fused[i].ChunkID.String() < fused[j].ChunkID.String()
		}
		return fused[i].Score > fused[j].Score
	})

	for i := range fused {
		fused[i].Rank = i + 1
	}

	return fused
}
