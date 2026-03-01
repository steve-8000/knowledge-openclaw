package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/openclaw/ki-db/internal/config"
	"github.com/openclaw/ki-db/internal/db"
	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/internal/outbox"
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

	relay := outbox.NewRelay(pool, nats, cfg.Outbox.BatchSize, cfg.Outbox.PollInterval, logger)
	if err := relay.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("outbox relay failed", "error", err)
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
