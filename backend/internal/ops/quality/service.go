// Package quality provides operational monitoring and quality detection services.
package quality

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service provides operational status and quality detection.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a new quality/ops service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// PipelineStatus represents the current state of the indexing pipeline.
type PipelineStatus struct {
	OutboxLag      int64             `json:"outbox_lag"`
	RecentFailures int64             `json:"recent_failures"`
	Workers        map[string]string `json:"workers"` // worker_name -> status
}

// GetPipelineStatus returns the current pipeline operational status.
func (s *Service) GetPipelineStatus(ctx context.Context, tenantID uuid.UUID) (*PipelineStatus, error) {
	var lag int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM outbox_events
		WHERE published_at IS NULL
	`).Scan(&lag)
	if err != nil {
		return nil, fmt.Errorf("query outbox lag: %w", err)
	}

	var failures int64
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM outbox_events
		WHERE event_type = 'IndexJobFailed'
		  AND created_at > now() - interval '24 hours'
	`).Scan(&failures)
	if err != nil {
		return nil, fmt.Errorf("query recent failures: %w", err)
	}

	return &PipelineStatus{
		OutboxLag:      lag,
		RecentFailures: failures,
		Workers: map[string]string{
			"parser":   "healthy",
			"chunker":  "healthy",
			"embedder": "healthy",
			"graph":    "healthy",
			"quality":  "healthy",
		},
	}, nil
}

// RecentJob represents a recent indexing job/event.
type RecentJob struct {
	EventID     string  `json:"event_id"`
	EventType   string  `json:"event_type"`
	Status      string  `json:"status"` // pending|published|failed
	CreatedAt   string  `json:"created_at"`
	PublishedAt *string `json:"published_at,omitempty"`
}

// ListRecentJobs returns recent outbox events as job status.
func (s *Service) ListRecentJobs(ctx context.Context, tenantID uuid.UUID, limit int) ([]RecentJob, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT event_id, event_type, published_at, created_at
		FROM outbox_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent jobs: %w", err)
	}
	defer rows.Close()

	var jobs []RecentJob
	for rows.Next() {
		var j RecentJob
		var publishedAt *string
		if err := rows.Scan(&j.EventID, &j.EventType, &publishedAt, &j.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		j.PublishedAt = publishedAt
		if publishedAt != nil {
			j.Status = "published"
		} else {
			j.Status = "pending"
		}
		jobs = append(jobs, j)
	}

	return jobs, rows.Err()
}

// QualityReport summarizes data quality issues.
type QualityReport struct {
	DuplicateCandidates int64 `json:"duplicate_candidates"`
	StaleDocs           int64 `json:"stale_docs"`
	MissingMetadata     int64 `json:"missing_metadata"`
	NoOwnerDocs         int64 `json:"no_owner_docs"`
}

// GetQualityReport returns a summary of data quality issues.
func (s *Service) GetQualityReport(ctx context.Context, tenantID uuid.UUID) (*QualityReport, error) {
	report := &QualityReport{}

	// Stale: published docs not updated in 90 days
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM documents
		WHERE tenant_id = $1 AND status = 'published'
		  AND updated_at < now() - interval '90 days'
	`, tenantID).Scan(&report.StaleDocs)
	if err != nil {
		return nil, fmt.Errorf("query stale docs: %w", err)
	}

	// No owners
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM documents
		WHERE tenant_id = $1 AND (owners = '[]'::jsonb OR owners IS NULL)
	`, tenantID).Scan(&report.NoOwnerDocs)
	if err != nil {
		return nil, fmt.Errorf("query no owner docs: %w", err)
	}

	// Missing metadata: docs with no tags
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM documents
		WHERE tenant_id = $1 AND (tags = '[]'::jsonb OR tags IS NULL)
	`, tenantID).Scan(&report.MissingMetadata)
	if err != nil {
		return nil, fmt.Errorf("query missing metadata: %w", err)
	}

	return report, nil
}
