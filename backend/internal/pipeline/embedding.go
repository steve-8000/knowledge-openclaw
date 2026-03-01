package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmbeddingClient struct {
	httpClient *http.Client
	url        string
	apiKey     string
	model      string
	dims       int
}

func NewEmbeddingClient(url string, apiKey string, model string, dims int, timeout time.Duration) *EmbeddingClient {
	return &EmbeddingClient{
		httpClient: &http.Client{Timeout: timeout},
		url:        strings.TrimSpace(url),
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		dims:       dims,
	}
}

func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body := map[string]any{
		"model": c.model,
		"input": texts,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("build embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embedding provider: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding provider status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	vectors := make([][]float32, 0, len(texts))
	if len(parsed.Data) > 0 {
		for _, item := range parsed.Data {
			vectors = append(vectors, c.normalizeDims(item.Embedding))
		}
	} else if len(parsed.Embedding) > 0 {
		vectors = append(vectors, c.normalizeDims(parsed.Embedding))
	}

	if len(vectors) != len(texts) {
		return nil, fmt.Errorf("embedding size mismatch: got %d vectors for %d texts", len(vectors), len(texts))
	}

	return vectors, nil
}

func (c *EmbeddingClient) normalizeDims(vector []float32) []float32 {
	out := make([]float32, c.dims)
	copy(out, vector)
	return out
}
