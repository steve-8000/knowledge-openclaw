'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { CheckCircle2, XCircle } from 'lucide-react';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, listDocuments, apiFetch } from '@/lib/api/client';
import type { Document } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

export function CurationQueue(): JSX.Element {
  const [queue, setQueue] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pendingDocId, setPendingDocId] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    const tenantId = getTenantId();

    setLoading(true);
    setError(null);

    listDocuments(tenantId)
      .then((documents) => {
        if (!mounted) {
          return;
        }
        setQueue(documents.filter((doc) => doc.status === 'inbox'));
      })
      .catch((err: unknown) => {
        if (!mounted) {
          return;
        }
        setQueue([]);
        setError(err instanceof Error ? err.message : 'Failed to load curation queue');
      })
      .finally(() => {
        if (mounted) {
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, []);

  const updateStatus = async (doc: Document, status: 'published' | 'deprecated'): Promise<void> => {
    const tenantId = getTenantId();
    setPendingDocId(doc.doc_id);
    setError(null);

    try {
      await apiFetch<{ doc_id: string; status: string }>('ingest', `/api/v1/documents/${doc.doc_id}/status`, {
        tenantId,
        method: 'PATCH',
        body: JSON.stringify({ status })
      });
      setQueue((current) => current.filter((item) => item.doc_id !== doc.doc_id));
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update document status');
    } finally {
      setPendingDocId(null);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Review Queue</CardTitle>
        <CardDescription>Approve to publish into canonical index, reject to request fixes in upstream ingest metadata.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {error ? <p className="rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">{error}</p> : null}
        {loading ? (
          <p className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">Loading queue...</p>
        ) : null}
        {!loading && queue.length === 0 ? (
          <p className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">Queue empty. No pending curation actions.</p>
        ) : (
          queue.map((item) => (
            <div key={item.doc_id} className="flex flex-col gap-3 rounded-lg border p-4 md:flex-row md:items-center md:justify-between">
              <div>
                <Link href={`/documents/${item.doc_id}`} className="font-medium underline-offset-4 hover:underline">
                  {item.title}
                </Link>
                <p className="text-sm text-muted-foreground">Type: {item.doc_type}</p>
                <Badge variant="outline" className="mt-2">
                  Confidence {item.confidence}
                </Badge>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" disabled={pendingDocId === item.doc_id} onClick={() => void updateStatus(item, 'deprecated')}>
                  <XCircle className="mr-2 h-4 w-4" />
                  Reject
                </Button>
                <Button disabled={pendingDocId === item.doc_id} onClick={() => void updateStatus(item, 'published')}>
                  <CheckCircle2 className="mr-2 h-4 w-4" />
                  Approve
                </Button>
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}
