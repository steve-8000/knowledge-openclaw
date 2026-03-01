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
	"github.com/openclaw/ki-db/internal/outbox"
	"github.com/openclaw/ki-db/internal/pipeline"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/internal/workers"
	"github.com/openclaw/ki-db/pkg/events"
	"github.com/openclaw/ki-db/pkg/idempotency"
	"github.com/openclaw/ki-db/pkg/models"
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
		ctx := tenancy.WithTenant(parent, env.TenantID)
		tx, err := tenancy.BeginTx(ctx, pool)
		if err != nil {
			return fmt.Errorf("begin tenant tx: %w", err)
		}
		defer tx.Rollback(ctx)

		workerName := "worker-graph-" + string(env.EventType)
		alreadyProcessed, err := store.CheckAndMark(ctx, tx, env.EventID, workerName)
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return tx.Commit(ctx)
		}

		docID, links, err := resolveLinks(ctx, tx, env)
		if err != nil {
			return err
		}

		deleteTag, err := tx.Exec(ctx, `
			DELETE FROM edges
			WHERE from_doc_id = $1 AND relation = $2
		`, docID, models.RelationReferences)
		if err != nil {
			return fmt.Errorf("delete old graph edges: %w", err)
		}
		edgesRemoved := int(deleteTag.RowsAffected())

		edgesCreated, err := createEdges(ctx, tx, docID, links)
		if err != nil {
			return err
		}

		if err := writer.Write(ctx, tx, env.TenantID, events.TypeGraphUpdated, events.GraphUpdatedPayload{
			DocID:        docID,
			EdgesCreated: edgesCreated,
			EdgesRemoved: edgesRemoved,
			Relations:    []string{models.RelationReferences},
		}); err != nil {
			return fmt.Errorf("write outbox GraphUpdated: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit graph tx: %w", err)
		}

		logger.Info("graph updated", "doc_id", docID, "edges_created", edgesCreated, "edges_removed", edgesRemoved)
		return nil
	}

	parsedConsumer, err := workers.NewConsumer(
		ctx,
		nats,
		"graph-parsed",
		events.TypeDocumentParsed.Subject(),
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create graph parsed consumer", "error", err)
		os.Exit(1)
	}

	chunkConsumer, err := workers.NewConsumer(
		ctx,
		nats,
		"graph-chunks",
		events.TypeChunksCreated.Subject(),
		handler,
		logger,
	)
	if err != nil {
		logger.Error("create graph chunks consumer", "error", err)
		os.Exit(1)
	}

	errCh := make(chan error, 2)
	go func() { errCh <- parsedConsumer.Run(ctx) }()
	go func() { errCh <- chunkConsumer.Run(ctx) }()

	err = <-errCh
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("graph worker failed", "error", err)
		os.Exit(1)
	}
}

func resolveLinks(ctx context.Context, tx pgx.Tx, env *events.Envelope) (uuid.UUID, []string, error) {
	switch env.EventType {
	case events.TypeDocumentParsed:
		var payload events.DocumentParsedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return uuid.Nil, nil, fmt.Errorf("decode DocumentParsed payload: %w", err)
		}
		if len(payload.Links) > 0 {
			return payload.DocID, payload.Links, nil
		}
		return payload.DocID, pipeline.ExtractMarkdownLinks(payload.NormalizedText), nil
	case events.TypeChunksCreated:
		var payload events.ChunksCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return uuid.Nil, nil, fmt.Errorf("decode ChunksCreated payload: %w", err)
		}
		var normalized string
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(normalized_text, '')
			FROM document_versions
			WHERE doc_id = $1 AND version_id = $2
		`, payload.DocID, payload.VersionID).Scan(&normalized); err != nil {
			return uuid.Nil, nil, fmt.Errorf("load normalized text for graph: %w", err)
		}
		return payload.DocID, pipeline.ExtractMarkdownLinks(normalized), nil
	default:
		return uuid.Nil, nil, fmt.Errorf("unsupported graph trigger: %s", env.EventType)
	}
}

func createEdges(ctx context.Context, tx pgx.Tx, fromDocID uuid.UUID, links []string) (int, error) {
	type edgeTarget struct {
		toDocID       *uuid.UUID
		toExternalKey string
		evidence      []byte
	}

	unique := make(map[string]edgeTarget)
	for _, link := range links {
		key := pipeline.StableKeyFromLink(link)
		if key == "" {
			continue
		}
		evidence, err := json.Marshal(map[string]any{"link": link})
		if err != nil {
			return 0, fmt.Errorf("marshal edge evidence: %w", err)
		}

		var toDocID uuid.UUID
		err = tx.QueryRow(ctx, `
			SELECT doc_id FROM documents WHERE stable_key = $1
		`, key).Scan(&toDocID)
		if err == nil {
			if toDocID == fromDocID {
				continue
			}
			copied := toDocID
			unique["doc:"+toDocID.String()] = edgeTarget{toDocID: &copied, evidence: evidence}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("lookup link target by stable_key %q: %w", key, err)
		}
		unique["external:"+key] = edgeTarget{toExternalKey: key, evidence: evidence}
	}

	edgesCreated := 0
	for _, target := range unique {
		_, err := tx.Exec(ctx, `
			INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, to_external_key, relation, evidence)
			VALUES (current_setting('app.tenant_id')::uuid, $1, $2, $3, $4, $5::jsonb)
		`, fromDocID, target.toDocID, nullIfEmpty(target.toExternalKey), models.RelationReferences, target.evidence)
		if err != nil {
			return edgesCreated, fmt.Errorf("insert edge: %w", err)
		}
		edgesCreated++
	}
	return edgesCreated, nil
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
