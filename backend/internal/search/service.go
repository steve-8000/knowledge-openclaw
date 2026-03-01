package search

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/openclaw/ki-db/internal/config"
	packbuilder "github.com/openclaw/ki-db/internal/search/context"
	"github.com/openclaw/ki-db/internal/search/rank"
	"github.com/openclaw/ki-db/internal/search/retrieval"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/pkg/models"
)

type SearchFilters struct {
	DocType string
	Status  string
	Tags    []string
}

type SearchOpts struct {
	TopK          int
	RerankEnabled bool
	Filters       SearchFilters
}

type Service struct {
	pool        *pgxpool.Pool
	searchCfg   config.SearchConfig
	rerankerCfg config.RerankerConfig
	reranker    rank.Reranker
}

func NewService(pool *pgxpool.Pool, searchCfg config.SearchConfig, rerankerCfg config.RerankerConfig, reranker rank.Reranker) *Service {
	if reranker == nil {
		reranker = rank.NoopReranker{}
	}
	return &Service{
		pool:        pool,
		searchCfg:   searchCfg,
		rerankerCfg: rerankerCfg,
		reranker:    reranker,
	}
}

func (s *Service) Search(ctx context.Context, query string, embedding []float32, opts SearchOpts) (*models.ContextPack, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("search service is not initialized")
	}
	if query == "" {
		return nil, fmt.Errorf("query must not be empty")
	}

	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant from context: %w", err)
	}

	bm25TopK, annTopK, outputLimit := s.resolveTopK(opts)

	var (
		bm25Results []retrieval.RankedChunk
		annResults  []retrieval.RankedChunk
		bm25Err     error
		annErr      error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		tx, txErr := tenancy.BeginTx(ctx, s.pool)
		if txErr != nil {
			bm25Err = fmt.Errorf("begin bm25 tx: %w", txErr)
			return
		}
		defer tx.Rollback(ctx)
		bm25Results, bm25Err = retrieval.SearchBM25(ctx, tx, query, tenantID, bm25TopK)
	}()

	go func() {
		defer wg.Done()
		tx, txErr := tenancy.BeginTx(ctx, s.pool)
		if txErr != nil {
			annErr = fmt.Errorf("begin ann tx: %w", txErr)
			return
		}
		defer tx.Rollback(ctx)
		annResults, annErr = retrieval.SearchANN(ctx, tx, embedding, tenantID, annTopK)
	}()

	wg.Wait()
	if bm25Err != nil {
		return nil, fmt.Errorf("bm25 retrieval failed: %w", bm25Err)
	}
	if annErr != nil {
		return nil, fmt.Errorf("ann retrieval failed: %w", annErr)
	}

	fused := rank.FuseRRF(bm25Results, annResults, s.searchCfg.RRFK)
	fused = applyFilters(fused, opts.Filters)

	if opts.RerankEnabled && s.rerankerCfg.Enabled {
		rerankTopN := s.searchCfg.RerankTopN
		if rerankTopN <= 0 || rerankTopN > len(fused) {
			rerankTopN = len(fused)
		}
		fused, err = s.reranker.Rerank(ctx, query, fused, rerankTopN)
		if err != nil {
			return nil, fmt.Errorf("rerank failed: %w", err)
		}
	}

	return packbuilder.BuildContextPack(query, fused, outputLimit), nil
}

func (s *Service) resolveTopK(opts SearchOpts) (bm25TopK, annTopK, outputLimit int) {
	bm25TopK = s.searchCfg.BM25TopK
	if bm25TopK <= 0 {
		bm25TopK = 200
	}
	annTopK = s.searchCfg.ANNTopK
	if annTopK <= 0 {
		annTopK = 200
	}

	if opts.TopK > 0 {
		bm25TopK = opts.TopK
		annTopK = opts.TopK
		outputLimit = opts.TopK
		return
	}

	if bm25TopK > annTopK {
		outputLimit = bm25TopK
		return
	}
	outputLimit = annTopK
	return
}

func applyFilters(in []retrieval.RankedChunk, filters SearchFilters) []retrieval.RankedChunk {
	if len(in) == 0 {
		return in
	}

	needsDocType := filters.DocType != ""
	needsStatus := filters.Status != ""
	needsTags := len(filters.Tags) > 0
	if !needsDocType && !needsStatus && !needsTags {
		return in
	}

	out := make([]retrieval.RankedChunk, 0, len(in))
	for _, chunk := range in {
		if needsDocType && chunk.DocType != filters.DocType {
			continue
		}
		if needsStatus && chunk.Status != filters.Status {
			continue
		}
		if needsTags && !hasAllTags(chunk.Tags, filters.Tags) {
			continue
		}
		out = append(out, chunk)
	}

	for i := range out {
		out[i].Rank = i + 1
	}

	return out
}

func hasAllTags(candidate, required []string) bool {
	if len(required) == 0 {
		return true
	}
	tagSet := make(map[string]struct{}, len(candidate))
	for _, tag := range candidate {
		tagSet[tag] = struct{}{}
	}
	for _, tag := range required {
		if _, ok := tagSet[tag]; !ok {
			return false
		}
	}
	return true
}
