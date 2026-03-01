// Package main provides a database migration runner for ki-db.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://kidb:kidb_secret@localhost:5432/kidb?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Determine direction
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	migrationsDir := findMigrationsDir()
	logger.Info("running migrations", "dir", migrationsDir, "direction", direction)

	if err := runMigrations(ctx, pool, migrationsDir, direction, logger); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed successfully")
}

func findMigrationsDir() string {
	// Check common locations
	candidates := []string{
		"db/migrations",
		"../db/migrations",
		"../../db/migrations",
		"/migrations",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return "db/migrations"
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir, direction string, logger *slog.Logger) error {
	// Create migrations tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Find migration files
	suffix := "." + direction + ".sql"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	if direction == "down" {
		// Reverse order for down migrations
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	for _, file := range files {
		version := strings.Split(file, "_")[0]

		if direction == "up" {
			// Check if already applied
			var exists bool
			err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
			if err != nil {
				return fmt.Errorf("check migration %s: %w", version, err)
			}
			if exists {
				logger.Info("skipping (already applied)", "version", version)
				continue
			}
		}

		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		logger.Info("applying migration", "file", file, "direction", direction)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", file, err)
		}

		if direction == "up" {
			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", version); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("record migration %s: %w", version, err)
			}
		} else {
			if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("remove migration record %s: %w", version, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}

		logger.Info("applied", "file", file)
	}

	return nil
}
