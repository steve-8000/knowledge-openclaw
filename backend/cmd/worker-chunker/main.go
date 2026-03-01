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

	"github.com/google/uuid"

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

	store := idempotency.NewStore()
	writer := outbox.NewWriter()

	handler := func(parent context.Context, env *events.Envelope) error {
		var payload events.DocumentParsedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("decode DocumentParsed payload: %w", err)
		}

		ctx := tenancy.WithTenant(parent, env.TenantID)
		tx, err := tenancy.BeginTx(ctx, pool)
		if err != nil {
			return fmt.Errorf("begin tenant tx: %w", err)
		}
		defer tx.Rollback(ctx)

		alreadyProcessed, err := store.CheckAndMark(ctx, tx, env.EventID, "worker-chunker")
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return tx.Commit(ctx)
		}

		var normalized string
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(normalized_text, '')
			FROM document_versions
			WHERE doc_id = $1 AND version_id = $2
		`, payload.DocID, payload.VersionID).Scan(&normalized); err != nil {
			return fmt.Errorf("load normalized text: %w", err)
		}

		chunks := pipeline.ChunkByHeadings(normalized, 512)
		if _, err := tx.Exec(ctx, `
			DELETE FROM chunks
			WHERE doc_id = $1 AND version_id = $2
		`, payload.DocID, payload.VersionID); err != nil {
			return fmt.Errorf("delete existing chunks: %w", err)
		}

		chunkIDs := make([]uuid.UUID, 0, len(chunks))
		for i, chunk := range chunks {
			var chunkID uuid.UUID
			if err := tx.QueryRow(ctx, `
				INSERT INTO chunks (tenant_id, doc_id, version_id, ordinal, heading_path, chunk_text, token_count, chunk_sha256)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING chunk_id
			`, env.TenantID, payload.DocID, payload.VersionID, i, nullIfEmpty(chunk.HeadingPath), chunk.Text, chunk.TokenCount, chunk.SHA256).Scan(&chunkID); err != nil {
				return fmt.Errorf("insert chunk %d: %w", i, err)
			}
			chunkIDs = append(chunkIDs, chunkID)
		}

		if err := writer.Write(ctx, tx, env.TenantID, events.TypeChunksCreated, events.ChunksCreatedPayload{
			DocID:     payload.DocID,
			VersionID: payload.VersionID,
			ChunkIDs:  chunkIDs,
			Count:     len(chunkIDs),
		}); err != nil {
			return fmt.Errorf("write outbox ChunksCreated: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit chunker tx: %w", err)
		}

		logger.Info("chunks created", "doc_id", payload.DocID, "version_id", payload.VersionID, "count", len(chunkIDs))
		return nil
	}

	consumer, err := workers.NewConsumer(
		ctx,
		nats,
		"chunker",
		events.TypeDocumentParsed.Subject(),
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create chunker consumer", "error", err)
		os.Exit(1)
	}

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("chunker worker failed", "error", err)
		os.Exit(1)
	}
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
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
