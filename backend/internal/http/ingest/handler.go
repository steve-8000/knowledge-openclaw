package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/openclaw/ki-db/internal/outbox"
	"github.com/openclaw/ki-db/internal/tenancy"
	"github.com/openclaw/ki-db/pkg/events"
	"github.com/openclaw/ki-db/pkg/models"
)

type Handler struct {
	pool   *pgxpool.Pool
	writer *outbox.Writer
	logger *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, writer *outbox.Writer, logger *slog.Logger) *Handler {
	return &Handler{pool: pool, writer: writer, logger: logger}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/documents", h.upsertDocument)
	r.Get("/documents", h.listDocuments)
	r.Get("/documents/{docID}", h.getDocument)
	r.Patch("/documents/{docID}/status", h.updateDocumentStatus)
}

type upsertDocumentRequest struct {
	StableKey     string         `json:"stable_key"`
	Title         string         `json:"title"`
	DocType       string         `json:"doc_type"`
	Status        string         `json:"status"`
	Confidence    string         `json:"confidence"`
	Owners        []string       `json:"owners"`
	Tags          []string       `json:"tags"`
	Source        map[string]any `json:"source"`
	ContentURI    string         `json:"content_uri"`
	RawText       string         `json:"raw_text"`
	ContentSHA256 string         `json:"content_sha256"`
}

type upsertDocumentResponse struct {
	DocID      uuid.UUID `json:"doc_id"`
	VersionID  uuid.UUID `json:"version_id"`
	VersionNo  int64     `json:"version_no"`
	StableKey  string    `json:"stable_key"`
	EventTypes []string  `json:"event_types"`
}

func (h *Handler) upsertDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req upsertDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}
	if strings.TrimSpace(req.StableKey) == "" {
		writeError(w, http.StatusBadRequest, errors.New("stable_key is required"))
		return
	}
	if strings.TrimSpace(req.DocType) == "" {
		writeError(w, http.StatusBadRequest, errors.New("doc_type is required"))
		return
	}

	if req.Status == "" {
		req.Status = models.StatusInbox
	}
	if req.Confidence == "" {
		req.Confidence = models.ConfidenceMed
	}
	if req.Owners == nil {
		req.Owners = []string{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.Source == nil {
		req.Source = map[string]any{}
	}

	contentHash, err := resolveContentHash(req.RawText, req.ContentSHA256)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	docID, err := h.upsertDocumentRow(ctx, tx, tenantID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	versionID, versionNo, err := h.insertVersion(ctx, tx, tenantID, docID, req, contentHash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.writer.Write(ctx, tx, tenantID, events.TypeDocumentUpserted, events.DocumentUpsertedPayload{
		DocID:     docID,
		StableKey: req.StableKey,
		Title:     req.Title,
		DocType:   req.DocType,
		Status:    req.Status,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("write document outbox event: %w", err))
		return
	}

	if err := h.writer.Write(ctx, tx, tenantID, events.TypeVersionCreated, events.VersionCreatedPayload{
		DocID:         docID,
		VersionID:     versionID,
		VersionNo:     versionNo,
		ContentURI:    req.ContentURI,
		ContentSHA256: hex.EncodeToString(contentHash),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("write version outbox event: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	writeJSON(w, http.StatusCreated, upsertDocumentResponse{
		DocID:      docID,
		VersionID:  versionID,
		VersionNo:  versionNo,
		StableKey:  req.StableKey,
		EventTypes: []string{string(events.TypeDocumentUpserted), string(events.TypeVersionCreated)},
	})
}

func (h *Handler) upsertDocumentRow(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, req upsertDocumentRequest) (uuid.UUID, error) {
	ownersJSON, err := json.Marshal(req.Owners)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal owners: %w", err)
	}
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal tags: %w", err)
	}
	sourceJSON, err := json.Marshal(req.Source)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal source: %w", err)
	}

	var docID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO documents (
			tenant_id, stable_key, title, doc_type, status, confidence, owners, tags, source
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9::jsonb)
		ON CONFLICT (tenant_id, stable_key)
		DO UPDATE SET
			title = EXCLUDED.title,
			doc_type = EXCLUDED.doc_type,
			status = EXCLUDED.status,
			confidence = EXCLUDED.confidence,
			owners = EXCLUDED.owners,
			tags = EXCLUDED.tags,
			source = EXCLUDED.source,
			updated_at = now()
		RETURNING doc_id
	`, tenantID, req.StableKey, req.Title, req.DocType, req.Status, req.Confidence, ownersJSON, tagsJSON, sourceJSON).Scan(&docID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert document: %w", err)
	}

	return docID, nil
}

func (h *Handler) insertVersion(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, docID uuid.UUID, req upsertDocumentRequest, contentHash []byte) (uuid.UUID, int64, error) {
	var versionNo int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(version_no), 0) + 1
		FROM document_versions
		WHERE tenant_id = $1 AND doc_id = $2
	`, tenantID, docID).Scan(&versionNo); err != nil {
		return uuid.Nil, 0, fmt.Errorf("resolve next version number: %w", err)
	}

	var versionID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO document_versions (
			tenant_id, doc_id, version_no, content_uri, raw_text, content_sha256
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING version_id
	`, tenantID, docID, versionNo, nullIfEmpty(req.ContentURI), nullIfEmpty(req.RawText), contentHash).Scan(&versionID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("insert document version: %w", err)
	}

	return versionID, versionNo, nil
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 500 {
			writeError(w, http.StatusBadRequest, errors.New("limit must be an integer between 1 and 500"))
			return
		}
		limit = parsed
	}

	rows, err := tx.Query(ctx, `
		SELECT tenant_id, doc_id, stable_key, COALESCE(title, ''), doc_type, status, confidence,
			owners, tags, source, created_at, updated_at
		FROM documents
		ORDER BY updated_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query documents: %w", err))
		return
	}
	defer rows.Close()

	docs := make([]models.Document, 0, limit)
	for rows.Next() {
		doc, err := scanDocumentRow(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("iterate documents: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"documents": docs})
}

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID, err := uuid.Parse(chi.URLParam(r, "docID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid docID: %w", err))
		return
	}

	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT tenant_id, doc_id, stable_key, COALESCE(title, ''), doc_type, status, confidence,
			owners, tags, source, created_at, updated_at
		FROM documents
		WHERE doc_id = $1
	`, docID)

	doc, err := scanDocumentRow(singleRow{row: row})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("document not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	verRows, err := tx.Query(ctx, `
		SELECT tenant_id, doc_id, version_id, version_no, COALESCE(content_uri, ''), COALESCE(raw_text, ''),
			COALESCE(normalized_text, ''), content_sha256, created_at
		FROM document_versions
		WHERE doc_id = $1
		ORDER BY version_no DESC
	`, docID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("query document versions: %w", err))
		return
	}
	defer verRows.Close()

	versions := make([]map[string]any, 0)
	for verRows.Next() {
		var v models.DocumentVersion
		if err := verRows.Scan(
			&v.TenantID,
			&v.DocID,
			&v.VersionID,
			&v.VersionNo,
			&v.ContentURI,
			&v.RawText,
			&v.NormalizedText,
			&v.ContentSHA256,
			&v.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("scan version: %w", err))
			return
		}
		versions = append(versions, map[string]any{
			"tenant_id":       v.TenantID,
			"doc_id":          v.DocID,
			"version_id":      v.VersionID,
			"version_no":      v.VersionNo,
			"content_uri":     v.ContentURI,
			"raw_text":        v.RawText,
			"normalized_text": v.NormalizedText,
			"content_sha256":  hex.EncodeToString(v.ContentSHA256),
			"created_at":      v.CreatedAt,
		})
	}
	if err := verRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("iterate versions: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"versions": versions,
	})
}

type scanner interface {
	Scan(dest ...any) error
}

type singleRow struct {
	row pgx.Row
}

func (s singleRow) Scan(dest ...any) error {
	return s.row.Scan(dest...)
}

func scanDocumentRow(row scanner) (models.Document, error) {
	var doc models.Document
	var ownersRaw []byte
	var tagsRaw []byte
	var sourceRaw []byte
	if err := row.Scan(
		&doc.TenantID,
		&doc.DocID,
		&doc.StableKey,
		&doc.Title,
		&doc.DocType,
		&doc.Status,
		&doc.Confidence,
		&ownersRaw,
		&tagsRaw,
		&sourceRaw,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	); err != nil {
		return models.Document{}, err
	}

	if len(ownersRaw) > 0 {
		if err := json.Unmarshal(ownersRaw, &doc.Owners); err != nil {
			return models.Document{}, fmt.Errorf("unmarshal owners: %w", err)
		}
	}
	if len(tagsRaw) > 0 {
		if err := json.Unmarshal(tagsRaw, &doc.Tags); err != nil {
			return models.Document{}, fmt.Errorf("unmarshal tags: %w", err)
		}
	}
	if len(sourceRaw) > 0 {
		if err := json.Unmarshal(sourceRaw, &doc.Source); err != nil {
			return models.Document{}, fmt.Errorf("unmarshal source: %w", err)
		}
	}

	if doc.Owners == nil {
		doc.Owners = []string{}
	}
	if doc.Tags == nil {
		doc.Tags = []string{}
	}
	if doc.Source == nil {
		doc.Source = map[string]any{}
	}

	return doc, nil
}

func resolveContentHash(rawText string, encoded string) ([]byte, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed != "" {
		decoded, err := hex.DecodeString(trimmed)
		if err != nil {
			return nil, fmt.Errorf("content_sha256 must be hex: %w", err)
		}
		if len(decoded) != sha256.Size {
			return nil, fmt.Errorf("content_sha256 must be %d-byte hash", sha256.Size)
		}
		return decoded, nil
	}

	if strings.TrimSpace(rawText) == "" {
		return nil, errors.New("raw_text or content_sha256 is required")
	}
	sum := sha256.Sum256([]byte(rawText))
	return sum[:], nil
}

func nullIfEmpty(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

var validStatuses = map[string]bool{
	models.StatusInbox:      true,
	models.StatusPublished:  true,
	models.StatusDeprecated: true,
	models.StatusArchived:   true,
}

func (h *Handler) updateDocumentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := tenancy.FromContext(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "docID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid docID: %w", err))
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if !validStatuses[req.Status] {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid status %q: must be one of inbox, published, deprecated, archived", req.Status))
		return
	}

	tx, err := tenancy.BeginTx(ctx, h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("begin tenant tx: %w", err))
		return
	}
	defer tx.Rollback(ctx)

	var stableKey, title, docType string
	err = tx.QueryRow(ctx, `
		UPDATE documents SET status = $1, updated_at = now()
		WHERE doc_id = $2
		RETURNING stable_key, COALESCE(title, ''), doc_type
	`, req.Status, docID).Scan(&stableKey, &title, &docType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("document not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Errorf("update document status: %w", err))
		return
	}

	if err := h.writer.Write(ctx, tx, tenantID, events.TypeDocumentUpserted, events.DocumentUpsertedPayload{
		DocID:     docID,
		StableKey: stableKey,
		Title:     title,
		DocType:   docType,
		Status:    req.Status,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("write outbox event: %w", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("commit transaction: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"doc_id": docID,
		"status": req.Status,
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
