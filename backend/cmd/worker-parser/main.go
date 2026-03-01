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
		var payload events.VersionCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("decode VersionCreated payload: %w", err)
		}

		ctx := tenancy.WithTenant(parent, env.TenantID)
		tx, err := tenancy.BeginTx(ctx, pool)
		if err != nil {
			return fmt.Errorf("begin tenant tx: %w", err)
		}
		defer tx.Rollback(ctx)

		alreadyProcessed, err := store.CheckAndMark(ctx, tx, env.EventID, "worker-parser")
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return tx.Commit(ctx)
		}

		var rawText string
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(raw_text, '')
			FROM document_versions
			WHERE doc_id = $1 AND version_id = $2
		`, payload.DocID, payload.VersionID).Scan(&rawText); err != nil {
			return fmt.Errorf("load raw text: %w", err)
		}

		normalized := pipeline.NormalizeWhitespace(rawText)
		links := pipeline.ExtractMarkdownLinks(normalized)

		if _, err := tx.Exec(ctx, `
			UPDATE document_versions
			SET normalized_text = $1
			WHERE doc_id = $2 AND version_id = $3
		`, normalized, payload.DocID, payload.VersionID); err != nil {
			return fmt.Errorf("update normalized text: %w", err)
		}

		if err := writer.Write(ctx, tx, env.TenantID, events.TypeDocumentParsed, events.DocumentParsedPayload{
			DocID:          payload.DocID,
			VersionID:      payload.VersionID,
			NormalizedText: normalized,
			Links:          links,
		}); err != nil {
			return fmt.Errorf("write outbox DocumentParsed: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit parser tx: %w", err)
		}

		logger.Info("parsed version", "doc_id", payload.DocID, "version_id", payload.VersionID, "link_count", len(links))
		return nil
	}

	consumer, err := workers.NewConsumer(
		ctx,
		nats,
		"parser",
		events.TypeVersionCreated.Subject(),
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create parser consumer", "error", err)
		os.Exit(1)
	}

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("parser worker failed", "error", err)
		os.Exit(1)
	}
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
