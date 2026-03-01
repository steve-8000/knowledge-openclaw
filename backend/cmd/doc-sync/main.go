package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sethvargo/go-envconfig"

	"github.com/openclaw/ki-db/internal/docsync"
)

type Config struct {
	DocsDir      string        `env:"DOCS_DIR,default=/docs"`
	IngestAPIURL string        `env:"INGEST_API_URL,default=http://ingest-api:8080"`
	TenantID     string        `env:"TENANT_ID,default=00000000-0000-0000-0000-000000000001"`
	StateFile    string        `env:"STATE_FILE,default=/state/sync-state.json"`
	SyncInterval time.Duration `env:"SYNC_INTERVAL,default=60s"`
	RunOnce      bool          `env:"RUN_ONCE,default=false"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		logger.Error("config", "error", err)
		os.Exit(1)
	}

	logger.Info("doc-sync starting",
		"docs_dir", cfg.DocsDir,
		"api_url", cfg.IngestAPIURL,
		"interval", cfg.SyncInterval,
		"run_once", cfg.RunOnce,
	)

	syncer := docsync.NewSyncer(cfg.DocsDir, cfg.IngestAPIURL, cfg.TenantID, cfg.StateFile, logger)

	if cfg.RunOnce {
		n, err := syncer.RunOnce()
		if err != nil {
			logger.Error("sync failed", "error", err)
			os.Exit(1)
		}
		logger.Info("sync completed", "synced", n)
		return
	}

	// Daemon mode: periodic sync
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initial sync
	n, err := syncer.RunOnce()
	if err != nil {
		logger.Error("initial sync failed", "error", err)
	} else {
		logger.Info("initial sync completed", "synced", n)
	}

	ticker := time.NewTicker(cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down")
			return
		case <-ticker.C:
			n, err := syncer.RunOnce()
			if err != nil {
				logger.Error("sync failed", "error", err)
			} else if n > 0 {
				logger.Info("sync completed", "synced", n)
			}
		}
	}
}
