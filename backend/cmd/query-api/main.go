package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/openclaw/ki-db/internal/config"
	"github.com/openclaw/ki-db/internal/db"
	queryhttp "github.com/openclaw/ki-db/internal/http/query"
	graphhttp "github.com/openclaw/ki-db/internal/http/graph"
	kihttp "github.com/openclaw/ki-db/internal/http"
	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/internal/outbox"
	"github.com/openclaw/ki-db/internal/search"
	"github.com/openclaw/ki-db/internal/search/rank"
	"github.com/openclaw/ki-db/internal/tenancy"
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

	var reranker rank.Reranker
	if cfg.Reranker.Enabled {
		reranker = rank.NewHTTPReranker(cfg.Reranker.URL, cfg.Reranker.Model)
	}
	searchSvc := search.NewService(pool, cfg.Search, cfg.Reranker, reranker)
	outboxWriter := outbox.NewWriter()
	queryHandler := queryhttp.NewHandler(searchSvc, pool, outboxWriter)
	graphHandler := graphhttp.NewHandler(pool)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(kihttp.CORS)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Route("/api/v1", func(api chi.Router) {
		api.Use(tenancy.Middleware(cfg.Tenancy.TenantHeader))
		queryHandler.Register(api)
		graphHandler.Register(api)
	})

	server := &http.Server{
		Addr:              ":8081",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("query api listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("query api failed", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "error", err)
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
