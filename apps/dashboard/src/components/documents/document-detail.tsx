'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import { AlertCircle, ArrowLeft, FileText } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, getDocument } from '@/lib/api/client';
import type { DocumentDetailResponse, DocumentVersion } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

type DocumentDetailProps = {
  docId: string;
};

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

function latestVersion(versions: DocumentVersion[]): DocumentVersion | null {
  if (versions.length === 0) {
    return null;
  }
  return [...versions].sort((a, b) => b.version_no - a.version_no)[0] ?? null;
}

function sortVersionsDesc(versions: DocumentVersion[]): DocumentVersion[] {
  return [...versions].sort((a, b) => b.version_no - a.version_no);
}

export function DocumentDetail({ docId }: DocumentDetailProps): JSX.Element {
  const [data, setData] = useState<DocumentDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedVersionId, setSelectedVersionId] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    const tenantId = getTenantId();

    setLoading(true);
    setError(null);

    getDocument(docId, tenantId)
      .then((response) => {
        if (!mounted) {
          return;
        }
        setData(response);
        const latest = latestVersion(response.versions ?? []);
        setSelectedVersionId(latest?.version_id ?? null);
      })
      .catch((err: unknown) => {
        if (!mounted) {
          return;
        }
        setData(null);
        setSelectedVersionId(null);
        setError(err instanceof Error ? err.message : 'Failed to load document');
      })
      .finally(() => {
        if (mounted) {
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, [docId]);

  const versions = useMemo(() => sortVersionsDesc(data?.versions ?? []), [data?.versions]);
  const currentVersion = useMemo(() => {
    if (versions.length === 0) {
      return null;
    }
    if (!selectedVersionId) {
      return versions[0];
    }
    return versions.find((v) => v.version_id === selectedVersionId) ?? versions[0];
  }, [versions, selectedVersionId]);
  const rawText = (currentVersion?.raw_text ?? '').trim();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <div className="space-y-1">
          <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Document</h1>
          <p className="text-muted-foreground">Open and inspect full document content from indexed versions.</p>
        </div>
        <Link href="/search" className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 py-2 text-sm text-foreground hover:bg-white/5">
          <ArrowLeft className="h-4 w-4" />
          Back to Search
        </Link>
      </div>

      {loading ? (
        <Card>
          <CardContent className="p-6 text-sm text-muted-foreground">Loading document...</CardContent>
        </Card>
      ) : null}

      {!loading && error ? (
        <Card>
          <CardContent className="flex items-center gap-2 p-6 text-sm text-destructive">
            <AlertCircle className="h-4 w-4" />
            {error}
          </CardContent>
        </Card>
      ) : null}

      {!loading && !error && data ? (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <FileText className="h-5 w-5 text-primary" />
                {data.document.title}
              </CardTitle>
              <CardDescription>{data.document.stable_key}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline">Type: {data.document.doc_type}</Badge>
                <Badge variant="outline">Status: {data.document.status}</Badge>
                <Badge variant="outline">Confidence: {data.document.confidence}</Badge>
                <Badge variant="outline">Versions: {data.versions.length}</Badge>
              </div>
              <div className="text-sm text-muted-foreground">
                <p>Doc ID: {data.document.doc_id}</p>
                <p>Updated: {new Date(data.document.updated_at).toLocaleString()}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Document Content</CardTitle>
              <CardDescription>
                {currentVersion ? `Version #${currentVersion.version_no}` : 'No version metadata found'}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {versions.length > 0 ? (
                <div className="max-w-sm">
                  <Select value={currentVersion?.version_id} onValueChange={(value) => setSelectedVersionId(value)}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select version" />
                    </SelectTrigger>
                    <SelectContent>
                      {versions.map((version) => (
                        <SelectItem key={version.version_id} value={version.version_id}>
                          v{version.version_no} - {new Date(version.created_at).toLocaleString()}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              ) : null}

              {rawText ? (
                <article className="rounded-lg border border-white/10 bg-black/20 p-4 text-sm leading-6 text-foreground/90">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={{
                      h1: ({ children }) => <h1 className="mb-3 text-2xl font-semibold">{children}</h1>,
                      h2: ({ children }) => <h2 className="mb-2 mt-5 text-xl font-semibold">{children}</h2>,
                      h3: ({ children }) => <h3 className="mb-2 mt-4 text-lg font-semibold">{children}</h3>,
                      p: ({ children }) => <p className="mb-3">{children}</p>,
                      ul: ({ children }) => <ul className="mb-3 list-disc pl-6">{children}</ul>,
                      ol: ({ children }) => <ol className="mb-3 list-decimal pl-6">{children}</ol>,
                      code: ({ children }) => <code className="rounded bg-white/10 px-1 py-0.5 text-xs">{children}</code>,
                      pre: ({ children }) => <pre className="mb-3 overflow-x-auto rounded bg-black/30 p-3 text-xs">{children}</pre>,
                      table: ({ children }) => <table className="mb-3 w-full border-collapse text-xs">{children}</table>,
                      thead: ({ children }) => <thead className="bg-white/10">{children}</thead>,
                      th: ({ children }) => <th className="border border-white/15 px-2 py-1 text-left">{children}</th>,
                      td: ({ children }) => <td className="border border-white/10 px-2 py-1 align-top">{children}</td>
                    }}
                  >
                    {rawText}
                  </ReactMarkdown>
                </article>
              ) : (
                <p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
                  This document has no stored raw text in the selected version.
                </p>
              )}
            </CardContent>
          </Card>
        </>
      ) : null}
    </div>
  );
}
