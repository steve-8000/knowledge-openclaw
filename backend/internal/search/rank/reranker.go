package rank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/openclaw/ki-db/internal/search/retrieval"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, chunks []retrieval.RankedChunk, topN int) ([]retrieval.RankedChunk, error)
}

type HTTPReranker struct {
	URL    string
	Model  string
	Client *http.Client
}

type rerankRequest struct {
	Query string   `json:"query"`
	Texts []string `json:"texts"`
	Model string   `json:"model,omitempty"`
}

type rerankResponse struct {
	Scores  []float64 `json:"scores"`
	Results []struct {
		Score float64 `json:"score"`
	} `json:"results"`
}

func NewHTTPReranker(url, model string) *HTTPReranker {
	return &HTTPReranker{
		URL:   url,
		Model: model,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *HTTPReranker) Rerank(ctx context.Context, query string, chunks []retrieval.RankedChunk, topN int) ([]retrieval.RankedChunk, error) {
	if len(chunks) == 0 || topN <= 0 {
		return chunks, nil
	}
	if r == nil || r.URL == "" {
		return nil, fmt.Errorf("reranker url is not configured")
	}
	if topN > len(chunks) {
		topN = len(chunks)
	}

	texts := make([]string, topN)
	for i := 0; i < topN; i++ {
		texts[i] = chunks[i].ChunkText
	}

	body, err := json.Marshal(rerankRequest{Query: query, Texts: texts, Model: r.Model})
	if err != nil {
		return nil, fmt.Errorf("marshal reranker request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build reranker request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call reranker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("reranker returned status %d: %s", resp.StatusCode, string(payload))
	}

	var parsed rerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode reranker response: %w", err)
	}

	scores := parsed.Scores
	if len(scores) == 0 && len(parsed.Results) > 0 {
		scores = make([]float64, len(parsed.Results))
		for i := range parsed.Results {
			scores[i] = parsed.Results[i].Score
		}
	}
	if len(scores) != topN {
		return nil, fmt.Errorf("reranker score length mismatch: got %d, want %d", len(scores), topN)
	}

	reordered := make([]retrieval.RankedChunk, len(chunks))
	copy(reordered, chunks)

	lead := make([]retrieval.RankedChunk, topN)
	copy(lead, reordered[:topN])
	for i := range lead {
		s := scores[i]
		lead[i].RerankScore = &s
	}

	sort.SliceStable(lead, func(i, j int) bool {
		return *lead[i].RerankScore > *lead[j].RerankScore
	})

	copy(reordered, lead)
	for i := range reordered {
		reordered[i].Rank = i + 1
	}

	return reordered, nil
}

type NoopReranker struct{}

func (NoopReranker) Rerank(_ context.Context, _ string, chunks []retrieval.RankedChunk, _ int) ([]retrieval.RankedChunk, error) {
	return chunks, nil
}
