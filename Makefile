.PHONY: help up down restart logs migrate build test lint clean \
       build-ingest build-query build-relay build-workers \
       dashboard-dev dashboard-build

# Default
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Infrastructure
# =============================================================================

up: ## Start all services (Postgres, NATS, etc.)
	docker compose -f infra/compose.yaml up -d

down: ## Stop all services
	docker compose -f infra/compose.yaml down

restart: down up ## Restart all services

logs: ## Tail service logs
	docker compose -f infra/compose.yaml logs -f

ps: ## Show running services
	docker compose -f infra/compose.yaml ps

# =============================================================================
# Database
# =============================================================================

migrate: ## Run database migrations
	@echo "Running migrations..."
	cd backend && go run ./cmd/migrate/... up

migrate-down: ## Rollback last migration
	cd backend && go run ./cmd/migrate/... down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=add_foo)
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=add_foo" && exit 1)
	@touch db/migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).up.sql
	@touch db/migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).down.sql
	@echo "Created migration: db/migrations/$$(date +%Y%m%d%H%M%S)_$(NAME)"

# =============================================================================
# Backend Build
# =============================================================================

build: build-ingest build-query build-relay build-workers ## Build all Go binaries

build-ingest: ## Build ingest API
	cd backend && go build -o ../bin/ingest-api ./cmd/ingest-api

build-query: ## Build query API
	cd backend && go build -o ../bin/query-api ./cmd/query-api

build-relay: ## Build outbox relay
	cd backend && go build -o ../bin/outbox-relay ./cmd/outbox-relay

build-workers: ## Build all workers
	cd backend && go build -o ../bin/worker-parser ./cmd/worker-parser
	cd backend && go build -o ../bin/worker-chunker ./cmd/worker-chunker
	cd backend && go build -o ../bin/worker-embedder ./cmd/worker-embedder
	cd backend && go build -o ../bin/worker-graph ./cmd/worker-graph
	cd backend && go build -o ../bin/worker-quality ./cmd/worker-quality

# =============================================================================
# Backend Dev
# =============================================================================

test: ## Run all Go tests
	cd backend && go test -race -count=1 ./...

test-v: ## Run tests with verbose output
	cd backend && go test -race -v -count=1 ./...

lint: ## Lint Go code
	cd backend && golangci-lint run ./...

vet: ## Go vet
	cd backend && go vet ./...

tidy: ## Go mod tidy
	cd backend && go mod tidy

# =============================================================================
# Dashboard
# =============================================================================

dashboard-dev: ## Start Next.js dashboard in dev mode
	cd apps/dashboard && npm run dev

dashboard-build: ## Build dashboard for production
	cd apps/dashboard && npm run build

dashboard-install: ## Install dashboard dependencies
	cd apps/dashboard && npm install

# =============================================================================
# All-in-one Dev
# =============================================================================

dev: up ## Start full dev environment
	@echo "Infrastructure started. Run services individually:"
	@echo "  make dashboard-dev    - Start dashboard"
	@echo "  go run ./cmd/ingest-api  - Start ingest API"
	@echo "  go run ./cmd/query-api   - Start query API"

clean: ## Clean build artifacts
	rm -rf bin/
	cd apps/dashboard && rm -rf .next out node_modules
