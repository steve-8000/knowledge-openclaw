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
	"github.com/jackc/pgx/v5"

	"github.com/openclaw/ki-db/internal/config"
	"github.com/openclaw/ki-db/internal/db"
	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/internal/pipeline"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/internal/workers"
	"github.com/openclaw/ki-db/pkg/events"
	"github.com/openclaw/ki-db/pkg/idempotency"
)

type chunkText struct {
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

	store := idempotency.NewStore()

	handler := func(parent context.Context, env *events.Envelope) error {
		docID, versionID, err := extractEntityIDs(env)
		if err != nil {
			logger.Warn("quality worker skip event", "event_type", env.EventType, "reason", err)
			return nil
		}

		ctx := tenancy.WithTenant(parent, env.TenantID)
		tx, err := tenancy.BeginTx(ctx, pool)
		if err != nil {
			return fmt.Errorf("begin tenant tx: %w", err)
		}
		defer tx.Rollback(ctx)

		alreadyProcessed, err := store.CheckAndMark(ctx, tx, env.EventID, "worker-quality")
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return tx.Commit(ctx)
		}

		if versionID != nil {
			chunks, err := loadChunkTexts(ctx, tx, *docID, *versionID)
			if err != nil {
				return err
			}
			warnNearDuplicates(logger, env.TenantID, *docID, *versionID, chunks)
		}

		if docID != nil {
			if err := checkMissingMetadata(ctx, tx, logger, env.TenantID, *docID); err != nil {
				return err
			}
			if err := checkStaleDocument(ctx, tx, logger, env.TenantID, *docID); err != nil {
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit quality tx: %w", err)
		}
		return nil
	}

	consumer, err := workers.NewConsumer(
		ctx,
		nats,
		"quality",
		"idx.doc.>",
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create quality consumer", "error", err)
		os.Exit(1)
	}

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("quality worker failed", "error", err)
		os.Exit(1)
	}
}

func extractEntityIDs(env *events.Envelope) (*uuid.UUID, *uuid.UUID, error) {
	switch env.EventType {
	case events.TypeVersionCreated:
		var payload events.VersionCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return nil, nil, err
		}
		return &payload.DocID, &payload.VersionID, nil
	case events.TypeDocumentParsed:
		var payload events.DocumentParsedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return nil, nil, err
		}
		return &payload.DocID, &payload.VersionID, nil
	case events.TypeChunksCreated:
		var payload events.ChunksCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return nil, nil, err
		}
		return &payload.DocID, &payload.VersionID, nil
	case events.TypeEmbeddingsGenerated:
		var payload events.EmbeddingsGeneratedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return nil, nil, err
		}
		return &payload.DocID, &payload.VersionID, nil
	default:
		return nil, nil, fmt.Errorf("unsupported event type: %s", env.EventType)
	}
}

func loadChunkTexts(ctx context.Context, tx pgx.Tx, docID uuid.UUID, versionID uuid.UUID) ([]chunkText, error) {
	rows, err := tx.Query(ctx, `
		SELECT chunk_id, chunk_text
		FROM chunks
		WHERE doc_id = $1 AND version_id = $2
		ORDER BY ordinal ASC
	`, docID, versionID)
	if err != nil {
		return nil, fmt.Errorf("query chunks for quality: %w", err)
	}
	defer rows.Close()

	out := make([]chunkText, 0)
	for rows.Next() {
		var item chunkText
		if err := rows.Scan(&item.ChunkID, &item.Text); err != nil {
			return nil, fmt.Errorf("scan chunk quality row: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quality chunks: %w", err)
	}
	return out, nil
}

func warnNearDuplicates(logger *slog.Logger, tenantID uuid.UUID, docID uuid.UUID, versionID uuid.UUID, chunks []chunkText) {
	for i := 0; i < len(chunks); i++ {
		for j := i + 1; j < len(chunks); j++ {
			score := pipeline.SimilarityJaccard(chunks[i].Text, chunks[j].Text)
			if score < 0.93 {
				continue
			}
			logger.Warn("potential duplicate chunks",
				"tenant_id", tenantID,
				"doc_id", docID,
				"version_id", versionID,
				"chunk_a", chunks[i].ChunkID,
				"chunk_b", chunks[j].ChunkID,
				"similarity", score,
			)
		}
	}
}

func checkMissingMetadata(ctx context.Context, tx pgx.Tx, logger *slog.Logger, tenantID uuid.UUID, docID uuid.UUID) error {
	var stableKey string
	var tagsCount int
	var ownersCount int
	err := tx.QueryRow(ctx, `
		SELECT stable_key, jsonb_array_length(tags), jsonb_array_length(owners)
		FROM documents
		WHERE doc_id = $1
	`, docID).Scan(&stableKey, &tagsCount, &ownersCount)
	if err != nil {
		return fmt.Errorf("load metadata counts: %w", err)
	}

	if tagsCount == 0 || ownersCount == 0 {
		logger.Warn("document missing metadata",
			"tenant_id", tenantID,
			"doc_id", docID,
			"stable_key", stableKey,
			"tags_count", tagsCount,
			"owners_count", ownersCount,
		)
	}
	return nil
}

func checkStaleDocument(ctx context.Context, tx pgx.Tx, logger *slog.Logger, tenantID uuid.UUID, docID uuid.UUID) error {
	var stableKey string
	var stale bool
	err := tx.QueryRow(ctx, `
		SELECT stable_key, updated_at < now() - interval '180 days'
		FROM documents
		WHERE doc_id = $1
	`, docID).Scan(&stableKey, &stale)
	if err != nil {
		return fmt.Errorf("load stale document status: %w", err)
	}

	if stale {
		logger.Warn("stale document detected", "tenant_id", tenantID, "doc_id", docID, "stable_key", stableKey)
	}
	return nil
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
