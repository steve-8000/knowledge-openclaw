package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"

	"github.com/openclaw/ki-db/internal/config"
	"github.com/openclaw/ki-db/internal/db"
	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/internal/outbox"
	"github.com/openclaw/ki-db/internal/pipeline"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/internal/workers"
	"github.com/openclaw/ki-db/pkg/events"
	"github.com/openclaw/ki-db/pkg/idempotency"
)

type chunkRecord struct {
	ChunkID uuid.UUID
	Text    string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	logger := newLogger(cfg.Log)

	pool, err := db.NewPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Error("create postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	nats, err := natsclient.NewClient(ctx, cfg.NATS)
	if err != nil {
		logger.Error("create nats client", "error", err)
		os.Exit(1)
	}
	defer nats.Close()

	embeddingURL := resolveEmbeddingURL(cfg.Embedding.URL, cfg.Embedding.Provider)
	embedClient := pipeline.NewEmbeddingClient(embeddingURL, cfg.Embedding.APIKey, cfg.Embedding.Model, cfg.Embedding.Dims, 45*time.Second)
	logger.Info("embedding provider configured", "provider", cfg.Embedding.Provider, "url", embeddingURL)

	store := idempotency.NewStore()
	writer := outbox.NewWriter()

	handler := func(parent context.Context, env *events.Envelope) error {
		var payload events.ChunksCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("decode ChunksCreated payload: %w", err)
		}
		if len(payload.ChunkIDs) == 0 {
			return nil
		}

		ctx := tenancy.WithTenant(parent, env.TenantID)
		tx, err := tenancy.BeginTx(ctx, pool)
		if err != nil {
			return fmt.Errorf("begin tenant tx: %w", err)
		}
		defer tx.Rollback(ctx)

		alreadyProcessed, err := store.CheckAndMark(ctx, tx, env.EventID, "worker-embedder")
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return tx.Commit(ctx)
		}

		records, err := loadChunks(ctx, tx, payload)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("commit empty embedder tx: %w", err)
			}
			return nil
		}

		for start := 0; start < len(records); start += cfg.Embedding.BatchSize {
			end := start + cfg.Embedding.BatchSize
			if end > len(records) {
				end = len(records)
			}
			batch := records[start:end]
			texts := make([]string, 0, len(batch))
			for _, item := range batch {
				texts = append(texts, item.Text)
			}

			vectors, err := embedClient.Embed(ctx, texts)
			if err != nil {
				return fmt.Errorf("embed batch [%d:%d]: %w", start, end, err)
			}

			for i, item := range batch {
				vec := pgvector.NewVector(vectors[i])
				if _, err := tx.Exec(ctx, `
					INSERT INTO chunk_embeddings (tenant_id, chunk_id, embedding_model, dims, embedding)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (tenant_id, chunk_id, embedding_model)
					DO UPDATE SET dims = EXCLUDED.dims, embedding = EXCLUDED.embedding, created_at = now()
				`, env.TenantID, item.ChunkID, cfg.Embedding.Model, cfg.Embedding.Dims, vec); err != nil {
					return fmt.Errorf("upsert embedding for chunk %s: %w", item.ChunkID, err)
				}
			}
		}

		if err := writer.Write(ctx, tx, env.TenantID, events.TypeEmbeddingsGenerated, events.EmbeddingsGeneratedPayload{
			DocID:          payload.DocID,
			VersionID:      payload.VersionID,
			ChunkIDs:       payload.ChunkIDs,
			EmbeddingModel: cfg.Embedding.Model,
			Dims:           cfg.Embedding.Dims,
		}); err != nil {
			return fmt.Errorf("write outbox EmbeddingsGenerated: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit embedder tx: %w", err)
		}

		logger.Info("embeddings generated", "doc_id", payload.DocID, "version_id", payload.VersionID, "chunks", len(payload.ChunkIDs))
		return nil
	}

	consumer, err := workers.NewConsumer(
		ctx,
		nats,
		"embedder",
		events.TypeChunksCreated.Subject(),
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create embedder consumer", "error", err)
		os.Exit(1)
	}

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("embedder worker failed", "error", err)
		os.Exit(1)
	}
}

func loadChunks(ctx context.Context, tx pgx.Tx, payload events.ChunksCreatedPayload) ([]chunkRecord, error) {
	rows, err := tx.Query(ctx, `
		SELECT chunk_id, chunk_text
		FROM chunks
		WHERE doc_id = $1 AND version_id = $2
		ORDER BY ordinal ASC
	`, payload.DocID, payload.VersionID)
	if err != nil {
		return nil, fmt.Errorf("load chunks: %w", err)
	}
	defer rows.Close()

	records := make([]chunkRecord, 0, len(payload.ChunkIDs))
	for rows.Next() {
		var item chunkRecord
		if err := rows.Scan(&item.ChunkID, &item.Text); err != nil {
			return nil, fmt.Errorf("scan chunk row: %w", err)
		}
		records = append(records, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chunk rows: %w", err)
	}
	return records, nil
}

func newLogger(cfg config.LogConfig) *slog.Logger {
	level := new(slog.LevelVar)
	switch cfg.Level {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	if cfg.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// resolveEmbeddingURL returns the embedding API endpoint.
// Priority: explicit URL override > provider default.
func resolveEmbeddingURL(explicitURL, provider string) string {
	if explicitURL != "" {
		return explicitURL
	}
	switch provider {
	case "ollama":
		return "http://localhost:11434/api/embeddings"
	case "openai":
		return "https://api.openai.com/v1/embeddings"
	default:
		return "https://api.openai.com/v1/embeddings"
	}
}
