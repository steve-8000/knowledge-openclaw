'use client';

import { useEffect, useState } from 'react';
import { AlertTriangle, Copy, FileWarning } from 'lucide-react';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, apiFetch } from '@/lib/api/client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

type QualityMetrics = {
  duplicate_candidates: number;
  stale_documents: number;
  missing_metadata: number;
};

const defaultMetrics: QualityMetrics = {
  duplicate_candidates: 0,
  stale_documents: 0,
  missing_metadata: 0
};

const metricCards = [
  {
    key: 'duplicate_candidates' as const,
    title: 'Duplicate Candidates',
    description: 'Potential duplicate document pairs with high semantic overlap.',
    icon: Copy
  },
  {
    key: 'stale_documents' as const,
    title: 'Stale Documents',
    description: 'Documents not updated within configured freshness threshold.',
    icon: FileWarning
  },
  {
    key: 'missing_metadata' as const,
    title: 'Missing Metadata',
    description: 'Documents lacking required ownership, tags, or source fields.',
    icon: AlertTriangle
  }
];

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

export function QualityDashboard(): JSX.Element {
  const [metrics, setMetrics] = useState<QualityMetrics>(defaultMetrics);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    setLoading(true);
    setError(null);

    apiFetch<QualityMetrics>('ingest', '/api/v1/ops/quality', { tenantId: getTenantId() })
      .then((response) => {
        if (!mounted) {
          return;
        }
        setMetrics(response);
      })
      .catch(() => {
        if (!mounted) {
          return;
        }
        setMetrics(defaultMetrics);
        setError('Unable to load quality metrics; displaying fallback values.');
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

  return (
    <div className="space-y-3">
      {error ? <p className="rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">{error}</p> : null}
      <div className="grid gap-4 md:grid-cols-3">
      {metricCards.map((metric) => {
        const Icon = metric.icon;
        return (
          <Card key={metric.title} className="overflow-hidden">
            <CardHeader className="pb-2">
              <CardDescription className="flex items-center gap-2 text-xs uppercase tracking-wide">
                <Icon className="h-4 w-4" />
                {metric.title}
              </CardDescription>
              <CardTitle className="text-3xl">{loading ? '-' : metrics[metric.key]}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">{metric.description}</p>
            </CardContent>
          </Card>
        );
      })}
      </div>
    </div>
  );
}
