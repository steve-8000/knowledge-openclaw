package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Postgres  PostgresConfig  `env:", prefix=POSTGRES_"`
	NATS      NATSConfig      `env:", prefix=NATS_"`
	Tenancy   TenancyConfig   `env:", prefix=TENANCY_"`
	Embedding EmbeddingConfig `env:", prefix=EMBEDDING_"`
	Reranker  RerankerConfig  `env:", prefix=RERANKER_"`
	Search    SearchConfig    `env:", prefix=SEARCH_"`
	Outbox    OutboxConfig    `env:", prefix=OUTBOX_"`
	Worker    WorkerConfig    `env:", prefix=WORKER_"`
	Log       LogConfig       `env:", prefix=LOG_"`
}

type PostgresConfig struct {
	DSN      string `env:"DSN, required"`
	MaxConns int32  `env:"MAX_CONNS, default=20"`
	MinConns int32  `env:"MIN_CONNS, default=2"`
}

type NATSConfig struct {
	URL            string `env:"URL, default=nats://localhost:4222"`
	Stream         string `env:"STREAM, default=IDX_DOCS"`      // via JETSTREAM_STREAM
	ConsumerPrefix string `env:"CONSUMER_PREFIX, default=kidb"` // via JETSTREAM_CONSUMER_PREFIX
}

type TenancyConfig struct {
	Mode         string `env:"MODE, default=rls"`
	TenantHeader string `env:"TENANT_HEADER, default=X-Tenant-ID"` // via TENANT_HEADER
}

type EmbeddingConfig struct {
	Provider  string `env:"PROVIDER, default=openai"`
	Model     string `env:"MODEL, default=text-embedding-3-small"`
	Dims      int    `env:"DIMS, default=1024"`
	APIKey    string `env:"API_KEY"`
	URL       string `env:"URL"` // explicit override; if empty, derived from Provider
	BatchSize int    `env:"BATCH_SIZE, default=100"`
}

type RerankerConfig struct {
	Enabled bool   `env:"ENABLED, default=false"`
	URL     string `env:"URL, default=http://localhost:8081/rerank"`
	Model   string `env:"MODEL, default=cross-encoder/ms-marco-MiniLM-L-6-v2"`
}

type SearchConfig struct {
	BM25TopK   int `env:"BM25_TOPK, default=200"`
	ANNTopK    int `env:"ANN_TOPK, default=200"`
	RerankTopN int `env:"RERANK_TOPN, default=50"`
	RRFK       int `env:"RRF_K, default=60"`
}

type OutboxConfig struct {
	BatchSize    int           `env:"BATCH_SIZE, default=100"`
	PollInterval time.Duration `env:"POLL_INTERVAL_MS, default=500ms"`
	MaxRetries   int           `env:"MAX_RETRIES, default=5"`
}

type WorkerConfig struct {
	MaxRetries    int           `env:"MAX_RETRIES, default=8"`
	AckWait       time.Duration `env:"ACK_WAIT_SEC, default=30s"`
	MaxAckPending int           `env:"MAX_ACK_PENDING, default=500"`
}

type LogConfig struct {
	Level  string `env:"LEVEL, default=info"`
	Format string `env:"FORMAT, default=json"`
}

// Load reads configuration from environment variables.
func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
