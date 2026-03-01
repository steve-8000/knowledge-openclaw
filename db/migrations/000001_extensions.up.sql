-- 000001_extensions.up.sql
-- Enable required Postgres extensions: pgvector + pg_search (ParadeDB BM25)

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_search;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
