// Package tenancy provides multi-tenant context management and HTTP middleware.
package tenancy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const tenantKey contextKey = "tenant_id"

// FromContext extracts tenant ID from context.
func FromContext(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(tenantKey)
	if v == nil {
		return uuid.Nil, fmt.Errorf("tenant_id not found in context")
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant_id has invalid type")
	}
	return id, nil
}

// MustFromContext extracts tenant ID from context, panics if not found.
// Use only after Middleware has validated the tenant header.
func MustFromContext(ctx context.Context) uuid.UUID {
	id, err := FromContext(ctx)
	if err != nil {
		panic("tenancy.MustFromContext: " + err.Error())
	}
	return id
}

// WithTenant stores tenant ID in context.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantKey, tenantID)
}

// Middleware extracts tenant ID from the configured header and injects it into context.
func Middleware(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(headerName)
			if raw == "" {
				http.Error(w, `{"error":"missing tenant header"}`, http.StatusBadRequest)
				return
			}
			tenantID, err := uuid.Parse(raw)
			if err != nil {
				http.Error(w, `{"error":"invalid tenant_id format"}`, http.StatusBadRequest)
				return
			}
			ctx := WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BeginTx starts a transaction with RLS tenant context set.
// All queries within this transaction are tenant-isolated.
func BeginTx(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, error) {
	tenantID, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	// Set session-local tenant context for RLS policies
	_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID.String())
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("set tenant context: %w", err)
	}

	return tx, nil
}
