export type DocStatus = 'inbox' | 'published' | 'deprecated' | 'archived';

export interface Document {
  tenant_id: string;
  doc_id: string;
  stable_key: string;
  title: string;
  doc_type: string;
  status: DocStatus;
  confidence: 'high' | 'med' | 'low';
  owners: string[];
  tags: string[];
  source: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface DocumentVersion {
  tenant_id: string;
  doc_id: string;
  version_id: string;
  version_no: number;
  content_uri?: string;
  raw_text?: string;
  normalized_text?: string;
  content_sha256: string;
  created_at: string;
}

export interface DocumentDetailResponse {
  document: Document;
  versions: DocumentVersion[];
}

export interface Citation {
  doc_id: string;
  version_id: string;
  title: string;
  heading_path?: string;
  chunk_id: string;
  relevance: 'keyword' | 'semantic' | 'both';
}

export interface SearchResult {
  chunk_id: string;
  doc_id: string;
  title: string;
  heading_path?: string;
  chunk_text: string;
  bm25_rank?: number;
  ann_rank?: number;
  rrf_score: number;
  rerank_score?: number;
}

export interface ContextPack {
  query: string;
  results: SearchResult[];
  citations: Citation[];
}

export interface Edge {
  id: string;
  source_id: string;
  target_id: string;
  relation_type: string;
  weight?: number;
  created_at?: string;
}

export interface GraphEgoResponse {
  center_doc_id: string;
  hops: number;
  documents: Document[];
  edges: Edge[];
}

export interface WorkerStatus {
  name: string;
  status: 'online' | 'degraded' | 'offline';
  queue_depth: number;
  last_heartbeat: string;
}

export interface PipelineStatus {
  outbox_lag_seconds: number;
  workers: WorkerStatus[];
  ingest_rate_per_minute: number;
  error_rate: number;
}

export interface IndexingJob {
  id: string;
  document_id?: string;
  status: 'queued' | 'running' | 'success' | 'failed' | 'retrying';
  retries: number;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface OpsJobsResponse {
  jobs: IndexingJob[];
}
