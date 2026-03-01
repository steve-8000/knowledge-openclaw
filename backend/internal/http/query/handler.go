package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/openclaw/ki-db/internal/outbox"
	"github.com/openclaw/ki-db/internal/search"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/pkg/events"
)

type Handler struct {
	search *search.Service
	pool   *pgxpool.Pool
	outbox *outbox.Writer
}

func NewHandler(searchSvc *search.Service, pool *pgxpool.Pool, outboxWriter *outbox.Writer) *Handler {
	return &Handler{search: searchSvc, pool: pool, outbox: outboxWriter}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/search", h.searchDocuments)
	r.Post("/feedback", h.submitFeedback)
}

func (h *Handler) searchDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, errors.New("query parameter q is required"))
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 200 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid limit %q", raw))
			return
		}
		limit = parsed
	}

	ctx = tenancy.WithTenant(ctx, tenantID)
	pack, err := h.search.Search(ctx, query, nil, search.SearchOpts{TopK: limit})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("search failed: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, pack)
}

type feedbackRequest struct {
	Query           string     `json:"query"`
	SelectedChunkID *uuid.UUID `json:"selected_chunk_id,omitempty"`
	Helpful         *bool      `json:"helpful,omitempty"`
	Note            string     `json:"note,omitempty"`
}

func (h *Handler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		writeError(w, http.StatusBadRequest, errors.New("query is required"))
		return
	}

	ctx = tenancy.WithTenant(ctx, tenantID)
	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	feedbackID := uuid.New()
	var chunkID *uuid.UUID
	if req.SelectedChunkID != nil {
		chunkID = req.SelectedChunkID
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO search_feedback (tenant_id, feedback_id, query, selected_chunk_id, helpful, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tenantID, feedbackID, req.Query, chunkID, req.Helpful, req.Note)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("insert feedback: %w", err))
		return
	}

	payload := events.FeedbackReceivedPayload{
		Query:   req.Query,
		Helpful: req.Helpful,
		Note:    req.Note,
	}
	if chunkID != nil {
		payload.SelectedChunkID = *chunkID
	}
	if err := h.outbox.Write(ctx, tx, tenantID, events.TypeFeedbackReceived, payload); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("write outbox: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit: %w", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"feedback_id": feedbackID.String()})
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
