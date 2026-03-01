import type { ContextPack, Document, DocumentDetailResponse, GraphEgoResponse, IndexingJob, OpsJobsResponse, PipelineStatus } from '@/lib/api/types';

export const DEFAULT_TENANT_ID = '00000000-0000-0000-0000-000000000001';
export const TENANT_STORAGE_KEY = 'ki-db:tenant-id';

const QUERY_API_BASE_URL = process.env.NEXT_PUBLIC_QUERY_API_URL ?? 'http://localhost:8081';
const INGEST_API_BASE_URL = process.env.NEXT_PUBLIC_INGEST_API_URL ?? 'http://localhost:8080';

type ApiBase = 'query' | 'ingest';

type ApiFetchOptions = RequestInit & {
  tenantId?: string;
  params?: Record<string, string | number | boolean | undefined>;
};

export interface QualityMetrics {
  duplicate_candidates: number;
  stale_documents: number;
  missing_metadata: number;
}

function resolveBaseUrl(base: ApiBase): string {
  return base === 'query' ? QUERY_API_BASE_URL : INGEST_API_BASE_URL;
}

function buildUrl(base: ApiBase, path: string, params?: ApiFetchOptions['params']): string {
  const url = new URL(path, resolveBaseUrl(base));
  if (!params) {
    return url.toString();
  }

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return;
    }
    url.searchParams.set(key, String(value));
  });

  return url.toString();
}

export async function apiFetch<T>(base: ApiBase, path: string, options: ApiFetchOptions = {}): Promise<T> {
  const { tenantId = DEFAULT_TENANT_ID, params, headers, ...init } = options;
  const response = await fetch(buildUrl(base, path, params), {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-Tenant-ID': tenantId,
      ...headers
    },
    cache: 'no-store'
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`API request failed (${response.status}): ${text || response.statusText}`);
  }

  return (await response.json()) as T;
}

export async function listDocuments(tenantId?: string): Promise<Document[]> {
  const data = await apiFetch<{ documents?: Document[] }>('ingest', '/api/v1/documents', { tenantId });
  return data.documents ?? [];
}

export function getDocument(documentId: string, tenantId?: string): Promise<DocumentDetailResponse> {
  return apiFetch<DocumentDetailResponse>('ingest', `/api/v1/documents/${documentId}`, { tenantId });
}

export function updateDocumentStatus(docId: string, status: string, tenantId?: string): Promise<{ doc_id: string; status: string }> {
  return apiFetch<{ doc_id: string; status: string }>('ingest', `/api/v1/documents/${docId}/status`, {
    tenantId,
    method: 'PATCH',
    body: JSON.stringify({ status })
  });
}

export function searchDocuments(
  query: string,
  options: {
    docType?: string;
    status?: string;
    confidence?: string;
    rerank?: boolean;
    limit?: number;
    tenantId?: string;
  }
): Promise<ContextPack> {
  return apiFetch<ContextPack>('query', '/api/v1/search', {
    tenantId: options.tenantId,
    params: {
      q: query,
      rerank: options.rerank ?? true,
      limit: options.limit ?? 20,
      doc_type: options.docType,
      status: options.status,
      confidence: options.confidence
    }
  });
}

export function getGraphEgo(docId: string, hops: number, relationType?: string, tenantId?: string): Promise<GraphEgoResponse> {
  return apiFetch<GraphEgoResponse>('query', '/api/v1/graph/ego', {
    tenantId,
    params: {
      doc_id: docId,
      hops,
      relation_type: relationType
    }
  });
}

export function getOpsStatus(tenantId?: string): Promise<PipelineStatus> {
  return apiFetch<PipelineStatus>('ingest', '/api/v1/ops/status', { tenantId });
}

export function getQualityMetrics(tenantId?: string): Promise<QualityMetrics> {
  return apiFetch<QualityMetrics>('ingest', '/api/v1/ops/quality', { tenantId });
}

export async function getOpsJobs(tenantId?: string): Promise<IndexingJob[]> {
  const data = await apiFetch<OpsJobsResponse>('ingest', '/api/v1/ops/jobs', { tenantId });
  return data.jobs ?? [];
}
