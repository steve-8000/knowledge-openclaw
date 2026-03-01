// Package graph provides HTTP handlers for the knowledge graph API.
package graph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	graphsvc "github.com/openclaw/ki-db/internal/graph"
	"github.com/openclaw/ki-db/internal/tenancy"
)

// Handler serves knowledge graph HTTP endpoints.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new graph HTTP handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// Register mounts graph routes on the given router.
func (h *Handler) Register(r chi.Router) {
	r.Get("/graph/ego", h.getEgoGraph)
}

// --- Response DTOs matching frontend GraphEgoResponse ---

type egoResponse struct {
	CenterDocID string         `json:"center_doc_id"`
	Hops        int            `json:"hops"`
	Documents   []docResponse  `json:"documents"`
	Edges       []edgeResponse `json:"edges"`
}

type docResponse struct {
	TenantID   uuid.UUID      `json:"tenant_id"`
	DocID      uuid.UUID      `json:"doc_id"`
	StableKey  string         `json:"stable_key"`
	Title      string         `json:"title"`
	DocType    string         `json:"doc_type"`
	Status     string         `json:"status"`
	Confidence string         `json:"confidence"`
	Owners     []string       `json:"owners"`
	Tags       []string       `json:"tags"`
	Source     map[string]any `json:"source"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type edgeResponse struct {
	ID           string  `json:"id"`
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Weight       float64 `json:"weight,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
}

func (h *Handler) getEgoGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	docIDStr := r.URL.Query().Get("doc_id")
	if docIDStr == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("doc_id query parameter is required"))
		return
	}
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid doc_id: %w", err))
		return
	}

	hops := 2
	if raw := r.URL.Query().Get("hops"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 3 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("hops must be 1-3, got %q", raw))
			return
		}
		hops = parsed
	}

	// Open tenant-scoped transaction for RLS
	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	ego, err := graphsvc.GetEgoGraph(ctx, tx, tenantID, docID, hops)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("get ego graph: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit: %w", err))
		return
	}

	// Map to frontend-expected response shape
	resp := mapEgoResponse(docID, hops, ego)
	writeJSON(w, http.StatusOK, resp)
}

func mapEgoResponse(centerDocID uuid.UUID, hops int, ego *graphsvc.EgoGraph) egoResponse {
	docs := make([]docResponse, 0, len(ego.Nodes))
	for _, d := range ego.Nodes {
		owners := d.Owners
		if owners == nil {
			owners = []string{}
		}
		tags := d.Tags
		if tags == nil {
			tags = []string{}
		}
		source := d.Source
		if source == nil {
			source = map[string]any{}
		}
		docs = append(docs, docResponse{
			TenantID:   d.TenantID,
			DocID:      d.DocID,
			StableKey:  d.StableKey,
			Title:      d.Title,
			DocType:    d.DocType,
			Status:     d.Status,
			Confidence: d.Confidence,
			Owners:     owners,
			Tags:       tags,
			Source:     source,
			CreatedAt:  d.CreatedAt,
			UpdatedAt:  d.UpdatedAt,
		})
	}

	edges := make([]edgeResponse, 0, len(ego.Edges))
	for _, e := range ego.Edges {
		targetID := ""
		if e.ToDocID != nil {
			targetID = e.ToDocID.String()
		}
		edges = append(edges, edgeResponse{
			ID:           e.EdgeID.String(),
			SourceID:     e.FromDocID.String(),
			TargetID:     targetID,
			RelationType: e.Relation,
			CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		})
	}

	return egoResponse{
		CenterDocID: centerDocID.String(),
		Hops:        hops,
		Documents:   docs,
		Edges:       edges,
	}
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
