'use client';

import { useEffect, useState } from 'react';
import { Activity, AlertTriangle, Gauge, Server } from 'lucide-react';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, getOpsStatus } from '@/lib/api/client';
import type { PipelineStatus as PipelineStatusType } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

const initialStatus: PipelineStatusType = {
  outbox_lag_seconds: 0,
  workers: [],
  ingest_rate_per_minute: 0,
  error_rate: 0
};

export function PipelineStatus(): JSX.Element {
  const [status, setStatus] = useState<PipelineStatusType>(initialStatus);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    const fetchStatus = async () => {
      try {
        const data = await getOpsStatus(getTenantId());
        if (mounted) {
          setStatus(data);
          setError(null);
        }
      } catch (err: unknown) {
        if (mounted) {
          setStatus(initialStatus);
          setError(err instanceof Error ? err.message : 'Unable to load pipeline status');
        }
      }
    };

    void fetchStatus();
    const timer = window.setInterval(fetchStatus, 12000);

    return () => {
      mounted = false;
      window.clearInterval(timer);
    };
  }, []);

  return (
    <section className="space-y-3">
      {error ? (
        <div className="flex items-center gap-2 rounded-lg border border-destructive/35 bg-destructive/5 p-3 text-sm text-destructive">
          <AlertTriangle className="h-4 w-4" />
          {error}
        </div>
      ) : null}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Gauge className="h-4 w-4" />
              Outbox Lag
            </CardDescription>
            <CardTitle className="text-2xl">{status.outbox_lag_seconds}s</CardTitle>
          </CardHeader>
          <CardContent className="text-xs text-muted-foreground">Time for event propagation from ingest to query index.</CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Ingest Rate
            </CardDescription>
            <CardTitle className="text-2xl">{status.ingest_rate_per_minute}/min</CardTitle>
          </CardHeader>
          <CardContent className="text-xs text-muted-foreground">Current indexing throughput over the recent window.</CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              Error Rate
            </CardDescription>
            <CardTitle className="text-2xl">{(status.error_rate * 100).toFixed(2)}%</CardTitle>
          </CardHeader>
          <CardContent className="text-xs text-muted-foreground">Worker and pipeline execution failure ratio.</CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Server className="h-5 w-5 text-primary" />
            Worker Fleet
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {status.workers.length === 0 ? (
            <p className="text-sm text-muted-foreground">No worker telemetry available.</p>
          ) : (
            status.workers.map((worker) => (
              <div key={worker.name} className="flex items-center justify-between rounded-lg border p-3 text-sm">
                <div>
                  <p className="font-medium">{worker.name}</p>
                  <p className="text-xs text-muted-foreground">Queue depth: {worker.queue_depth}</p>
                </div>
                <Badge
                  variant={
                    worker.status === 'online' ? 'default' : worker.status === 'degraded' ? 'secondary' : 'destructive'
                  }
                >
                  {worker.status}
                </Badge>
              </div>
            ))
          )}
        </CardContent>
      </Card>
    </section>
  );
}
