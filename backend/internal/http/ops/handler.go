package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/openclaw/ki-db/internal/tenancy"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/ops/status", h.getStatus)
	r.Get("/ops/jobs", h.listJobs)
	r.Get("/ops/quality", h.getQuality)
}

type workerStatus struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	QueueDepth    int    `json:"queue_depth"`
	LastHeartbeat string `json:"last_heartbeat"`
}

type pipelineStatusResponse struct {
	OutboxLagSeconds    float64        `json:"outbox_lag_seconds"`
	Workers             []workerStatus `json:"workers"`
	IngestRatePerMinute int64          `json:"ingest_rate_per_minute"`
	ErrorRate           float64        `json:"error_rate"`
}

func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	var outboxLagSeconds *float64
	err = tx.QueryRow(ctx, `
		SELECT EXTRACT(EPOCH FROM (now() - MIN(created_at)))
		FROM outbox_events
		WHERE published_at IS NULL
	`).Scan(&outboxLagSeconds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query outbox lag seconds: %w", err))
		return
	}

	var ingestRatePerMinute int64
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM outbox_events
		WHERE published_at > now() - interval '60 seconds'
	`).Scan(&ingestRatePerMinute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query ingest rate per minute: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	lag := 0.0
	if outboxLagSeconds != nil {
		lag = *outboxLagSeconds
	}

	writeJSON(w, http.StatusOK, pipelineStatusResponse{
		OutboxLagSeconds:    lag,
		Workers:             []workerStatus{},
		IngestRatePerMinute: ingestRatePerMinute,
		ErrorRate:           0.0,
	})
}

type indexingJob struct {
	ID          string     `json:"id"`
	DocumentID  *string    `json:"document_id,omitempty"`
	Status      string     `json:"status"`
	Retries     int        `json:"retries"`
	Error       *string    `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type opsJobsResponse struct {
	Jobs []indexingJob `json:"jobs"`
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 500 {
			writeError(w, http.StatusBadRequest, errors.New("limit must be an integer between 1 and 500"))
			return
		}
		limit = parsed
	}

	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT
			event_id::text,
			payload->>'doc_id' AS document_id,
			published_at,
			created_at
		FROM outbox_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query ops jobs: %w", err))
		return
	}
	defer rows.Close()

	jobs := make([]indexingJob, 0, limit)
	for rows.Next() {
		var job indexingJob
		if err := rows.Scan(&job.ID, &job.DocumentID, &job.CompletedAt, &job.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("scan ops job: %w", err))
			return
		}

		if job.CompletedAt == nil {
			job.Status = "queued"
		} else {
			job.Status = "success"
		}

		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("iterate ops jobs: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, opsJobsResponse{Jobs: jobs})
}

type qualityMetrics struct {
	DuplicateCandidates int `json:"duplicate_candidates"`
	StaleDocuments      int `json:"stale_documents"`
	MissingMetadata     int `json:"missing_metadata"`
}

func (h *Handler) getQuality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	// Stale documents: not updated in 30 days
	var stale int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM documents
		WHERE updated_at < now() - interval '30 days'
	`).Scan(&stale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query stale docs: %w", err))
		return
	}

	// Missing metadata: documents with empty owners AND empty tags
	var missingMeta int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM documents
		WHERE (owners = '[]'::jsonb OR owners IS NULL)
		  AND (tags = '[]'::jsonb OR tags IS NULL)
	`).Scan(&missingMeta)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query missing metadata: %w", err))
		return
	}

	// Duplicate candidates: documents sharing the same title (case-insensitive)
	var dupes int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT LOWER(title) FROM documents
			WHERE title IS NOT NULL AND title != ''
			GROUP BY LOWER(title)
			HAVING COUNT(*) > 1
		) sub
	`).Scan(&dupes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query duplicate candidates: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, qualityMetrics{
		DuplicateCandidates: dupes,
		StaleDocuments:      stale,
		MissingMetadata:     missingMeta,
	})
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":"encode response"}`, http.StatusInternalServerError)
	}
}
