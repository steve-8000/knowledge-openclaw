'use client';

import { useCallback, useEffect, useState } from 'react';
import { RefreshCw } from 'lucide-react';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, getOpsJobs } from '@/lib/api/client';
import type { IndexingJob } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

function statusVariant(status: IndexingJob['status']): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (status === 'success') {
    return 'default';
  }
  if (status === 'failed') {
    return 'destructive';
  }
  if (status === 'running' || status === 'retrying') {
    return 'secondary';
  }
  return 'outline';
}

export function JobTable(): JSX.Element {
  const [jobs, setJobs] = useState<IndexingJob[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getOpsJobs(getTenantId());
      setJobs(data);
    } catch {
      setJobs([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchJobs();
  }, [fetchJobs]);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0">
        <CardTitle className="text-lg">Recent Indexing Jobs</CardTitle>
        <Button variant="outline" size="sm" onClick={() => void fetchJobs()} disabled={loading}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Job ID</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Document</TableHead>
              <TableHead>Retries</TableHead>
              <TableHead>Error</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {jobs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  {loading ? 'Loading jobs...' : 'No jobs available.'}
                </TableCell>
              </TableRow>
            ) : (
              jobs.map((job) => (
                <TableRow key={job.id}>
                  <TableCell className="font-mono text-xs">{job.id}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant(job.status)}>{job.status}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{job.document_id ?? '-'}</TableCell>
                  <TableCell>{job.retries}</TableCell>
                  <TableCell className="max-w-[260px] truncate text-xs text-muted-foreground">{job.error ?? '-'}</TableCell>
                  <TableCell>{new Date(job.created_at).toLocaleString()}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
