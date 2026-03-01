# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Build stage
# ---------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

ARG SERVICE=ingest-api

WORKDIR /app

# Cache dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source
COPY backend/ .

# Build the target service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service ./cmd/${SERVICE}

# ---------------------------------------------------------------------------
# Runtime stage
# ---------------------------------------------------------------------------
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /service /usr/local/bin/service

# Copy migrations (needed by some services)
COPY db/migrations /migrations

EXPOSE 8080 8081

ENTRYPOINT ["/usr/local/bin/service"]
