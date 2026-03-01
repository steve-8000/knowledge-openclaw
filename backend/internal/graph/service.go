// Package graph provides knowledge graph query services.
package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/openclaw/ki-db/pkg/models"
)

// Querier is the common interface for pgx.Tx and pgxpool.Pool.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// EgoGraph represents the nodes and edges around a focal document.
type EgoGraph struct {
	Nodes []models.Document `json:"nodes"`
	Edges []models.Edge     `json:"edges"`
}

// GetEgoGraph returns the subgraph within N hops of the given document.
func GetEgoGraph(ctx context.Context, q Querier, tenantID, docID uuid.UUID, hops int) (*EgoGraph, error) {
	if hops < 1 {
		hops = 1
	}
	if hops > 3 {
		hops = 3
	}

	// Recursive CTE to find all documents within N hops
	rows, err := q.Query(ctx, `
		WITH RECURSIVE reachable AS (
			-- Seed: the focal document
			SELECT $2::uuid AS doc_id, 0 AS depth
			UNION
			-- Expand: follow edges in both directions
			SELECT
				CASE
					WHEN e.from_doc_id = r.doc_id THEN e.to_doc_id
					ELSE e.from_doc_id
				END AS doc_id,
				r.depth + 1 AS depth
			FROM reachable r
			JOIN edges e ON e.tenant_id = $1
				AND (e.from_doc_id = r.doc_id OR e.to_doc_id = r.doc_id)
				AND e.to_doc_id IS NOT NULL
			WHERE r.depth < $3
		)
		SELECT DISTINCT d.doc_id, d.stable_key, d.title, d.doc_type, d.status,
		       d.confidence, d.owners, d.tags, d.source, d.created_at, d.updated_at
		FROM reachable r
		JOIN documents d ON d.tenant_id = $1 AND d.doc_id = r.doc_id
	`, tenantID, docID, hops)
	if err != nil {
		return nil, fmt.Errorf("query ego graph nodes: %w", err)
	}
	defer rows.Close()

	var nodes []models.Document
	docIDs := make(map[uuid.UUID]bool)
	for rows.Next() {
		var d models.Document
		d.TenantID = tenantID
		if err := rows.Scan(&d.DocID, &d.StableKey, &d.Title, &d.DocType, &d.Status,
			&d.Confidence, &d.Owners, &d.Tags, &d.Source, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, d)
		docIDs[d.DocID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nodes: %w", err)
	}

	// If no nodes found, return early with just the empty graph
	if len(docIDs) == 0 {
		return &EgoGraph{Nodes: nodes, Edges: nil}, nil
	}

	// Fetch edges between the discovered nodes
	edgeRows, err := q.Query(ctx, `
		SELECT e.edge_id, e.from_doc_id, e.to_doc_id, COALESCE(e.to_external_key, '') AS to_external_key, e.relation,
		       e.evidence, e.created_at
		FROM edges e
		WHERE e.tenant_id = $1
		  AND e.from_doc_id = ANY($2)
		  AND (e.to_doc_id = ANY($2) OR e.to_doc_id IS NULL)
	`, tenantID, uuidSlice(docIDs))
	if err != nil {
		return nil, fmt.Errorf("query ego graph edges: %w", err)
	}
	defer edgeRows.Close()

	var edges []models.Edge
	for edgeRows.Next() {
		var e models.Edge
		e.TenantID = tenantID
		if err := edgeRows.Scan(&e.EdgeID, &e.FromDocID, &e.ToDocID, &e.ToExternalKey,
			&e.Relation, &e.Evidence, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err := edgeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate edges: %w", err)
	}

	return &EgoGraph{Nodes: nodes, Edges: edges}, nil
}

func uuidSlice(m map[uuid.UUID]bool) []uuid.UUID {
	s := make([]uuid.UUID, 0, len(m))
	for id := range m {
		s = append(s, id)
	}
	return s
}
