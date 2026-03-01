'use client';

import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { AlertCircle, Braces, Network, SearchX } from 'lucide-react';

import { searchDocuments, DEFAULT_TENANT_ID, TENANT_STORAGE_KEY } from '@/lib/api/client';
import type { SearchResult } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

function escapeRegExp(input: string): string {
  return input.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function highlightText(content: string, query: string): React.ReactNode {
  if (!query) {
    return content;
  }

  const regex = new RegExp(escapeRegExp(query), 'ig');
  const nodes: React.ReactNode[] = [];
  let start = 0;
  let match = regex.exec(content);

  while (match) {
    const matchIndex = match.index;
    const matchedText = match[0];

    if (matchIndex > start) {
      nodes.push(<span key={`text-${start}`}>{content.slice(start, matchIndex)}</span>);
    }

    nodes.push(
      <mark key={`match-${matchIndex}`} className="rounded bg-primary/20 px-1 text-primary">
        {matchedText}
      </mark>
    );

    start = matchIndex + matchedText.length;
    match = regex.exec(content);
  }

  if (start < content.length) {
    nodes.push(<span key={`text-tail-${start}`}>{content.slice(start)}</span>);
  }

  return nodes;
}

export function SearchResults(): JSX.Element {
  const searchParams = useSearchParams();
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const query = searchParams.get('query') ?? '';
  const docType = searchParams.get('doc_type') ?? undefined;
  const status = searchParams.get('status') ?? undefined;
  const confidence = searchParams.get('confidence') ?? undefined;
  const rerank = searchParams.get('rerank') !== 'false';

  const hasQuery = query.trim().length > 0;

  useEffect(() => {
    if (!hasQuery) {
      setResults([]);
      setError(null);
      return;
    }

    let mounted = true;
    const tenantId = getTenantId();

    setLoading(true);
    setError(null);

    searchDocuments(query, { docType, status, confidence, rerank, limit: 20, tenantId })
      .then((response) => {
        if (!mounted) {
          return;
        }
        setResults(response.results ?? []);
      })
      .catch((err: unknown) => {
        if (!mounted) {
          return;
        }
        setResults([]);
        setError(err instanceof Error ? err.message : 'Failed to fetch search results');
      })
      .finally(() => {
        if (mounted) {
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, [query, docType, status, confidence, rerank, hasQuery]);

  const subtitle = useMemo(() => {
    if (loading) {
      return 'Fetching top candidate chunks...';
    }
    if (hasQuery && !error) {
      return `${results.length} result(s)`;
    }
    return 'Run a query to inspect ranked chunks and citations.';
  }, [loading, hasQuery, error, results.length]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Network className="h-5 w-5 text-primary" />
          Search Results
        </CardTitle>
        <CardDescription>{subtitle}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!hasQuery ? (
          <div className="flex min-h-40 flex-col items-center justify-center rounded-lg border border-dashed text-center text-sm text-muted-foreground">
            <SearchX className="mb-2 h-5 w-5" />
            Enter a query to search indexed documents.
          </div>
        ) : null}

        {hasQuery && error ? (
          <div className="flex items-center gap-2 rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
            <AlertCircle className="h-4 w-4" />
            {error}
          </div>
        ) : null}

        {hasQuery && !loading && !error && results.length === 0 ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">No results for this query and filter combination.</div>
        ) : null}

        {results.map((result) => (
          <article key={`${result.doc_id}:${result.chunk_id}`} className="rounded-lg border border-white/[0.06] bg-card p-4 shadow-sm transition-colors hover:bg-white/[0.03]">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <Link href={`/documents/${result.doc_id}`} className="font-semibold underline-offset-4 hover:underline">
                {result.title}
              </Link>
              {typeof result.bm25_rank === 'number' ? <Badge variant="outline">BM25 #{result.bm25_rank}</Badge> : null}
              {typeof result.ann_rank === 'number' ? <Badge variant="outline">ANN #{result.ann_rank}</Badge> : null}
              <Badge>{result.rrf_score.toFixed(3)}</Badge>
            </div>
            <p className="max-w-none text-sm text-foreground/90">{highlightText(result.chunk_text, query)}</p>
            <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <Braces className="h-3.5 w-3.5" />
              <span>Chunk: {result.chunk_id}</span>
              <span>Doc: {result.doc_id}</span>
            </div>
          </article>
        ))}
      </CardContent>
    </Card>
  );
}
