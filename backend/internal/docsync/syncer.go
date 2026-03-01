package docsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// SyncState tracks content hashes to detect changes.
type SyncState map[string]string // stable_key → content_hash

// Syncer manages document synchronization from filesystem to ki-db API.
type Syncer struct {
	docsDir    string
	apiURL     string
	tenantID   string
	stateFile  string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewSyncer creates a new document syncer.
func NewSyncer(docsDir, apiURL, tenantID, stateFile string, logger *slog.Logger) *Syncer {
	return &Syncer{
		docsDir:   docsDir,
		apiURL:    apiURL,
		tenantID:  tenantID,
		stateFile: stateFile,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// RunOnce performs a single sync cycle: scan → diff → ingest.
// Returns the number of documents synced.
func (s *Syncer) RunOnce() (int, error) {
	// 1. Load previous state
	state, err := s.loadState()
	if err != nil {
		s.logger.Warn("failed to load state, starting fresh", "error", err)
		state = make(SyncState)
	}

	// 2. Scan documents
	docs, err := ScanDocs(s.docsDir)
	if err != nil {
		return 0, fmt.Errorf("scan docs: %w", err)
	}
	s.logger.Info("scanned documents", "total", len(docs))

	// 3. Diff and ingest changed documents
	synced := 0
	newState := make(SyncState, len(docs))

	for _, doc := range docs {
		newState[doc.StableKey] = doc.ContentHash

		prevHash, exists := state[doc.StableKey]
		if exists && prevHash == doc.ContentHash {
			continue // No change
		}

		action := "new"
		if exists {
			action = "updated"
		}

		if err := s.ingest(doc); err != nil {
			s.logger.Error("ingest failed", "stable_key", doc.StableKey, "error", err)
			// Keep old hash so we retry next cycle
			if exists {
				newState[doc.StableKey] = prevHash
			} else {
				delete(newState, doc.StableKey)
			}
			continue
		}

		s.logger.Info("synced", "stable_key", doc.StableKey, "action", action)
		synced++
	}

	// 4. Detect deleted files (in state but not on disk)
	for key := range state {
		if _, exists := newState[key]; !exists {
			s.logger.Info("file removed from docs/", "stable_key", key)
			// We don't auto-delete from ki-db — just log the removal
		}
	}

	// 5. Save new state
	if err := s.saveState(newState); err != nil {
		return synced, fmt.Errorf("save state: %w", err)
	}

	return synced, nil
}

type ingestRequest struct {
	StableKey  string            `json:"stable_key"`
	Title      string            `json:"title"`
	DocType    string            `json:"doc_type"`
	Status     string            `json:"status,omitempty"`
	Confidence string            `json:"confidence,omitempty"`
	Owners     []string          `json:"owners,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Source     map[string]string `json:"source,omitempty"`
	RawText    string            `json:"raw_text"`
}

func (s *Syncer) ingest(doc DocFile) error {
	req := ingestRequest{
		StableKey:  doc.StableKey,
		Title:      doc.Meta.Title,
		DocType:    doc.Meta.DocType,
		Confidence: doc.Meta.Confidence,
		Owners:     doc.Meta.Owners,
		Tags:       doc.Meta.Tags,
		Source:     doc.Meta.Source,
		RawText:    doc.Body,
	}

	// Use frontmatter status if set, otherwise default to "published"
	if doc.Meta.Status != "" {
		req.Status = doc.Meta.Status
	} else {
		req.Status = "published"
	}

	// Use title from frontmatter, fall back to stable_key
	if req.Title == "" {
		req.Title = doc.StableKey
	}

	// Default doc_type
	if req.DocType == "" {
		req.DocType = "other"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/documents", s.apiURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", s.tenantID)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *Syncer) loadState() (SyncState, error) {
	data, err := os.ReadFile(s.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(SyncState), nil
		}
		return nil, err
	}

	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func (s *Syncer) saveState(state SyncState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.stateFile, data, 0644)
}
