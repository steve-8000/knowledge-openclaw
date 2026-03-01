'use client';

import { FormEvent, useMemo, useState } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { Search, SlidersHorizontal } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';

export function SearchForm(): JSX.Element {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const currentValues = useMemo(
    () => ({
      query: searchParams.get('query') ?? '',
      docType: searchParams.get('doc_type') ?? 'all',
      status: searchParams.get('status') ?? 'all',
      confidence: searchParams.get('confidence') ?? 'all',
      rerank: searchParams.get('rerank') !== 'false'
    }),
    [searchParams]
  );

  const [query, setQuery] = useState(currentValues.query);
  const [docType, setDocType] = useState(currentValues.docType);
  const [status, setStatus] = useState(currentValues.status);
  const [confidence, setConfidence] = useState(currentValues.confidence);
  const [rerank, setRerank] = useState(currentValues.rerank);

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const params = new URLSearchParams(searchParams.toString());

    if (query.trim().length > 0) {
      params.set('query', query.trim());
    } else {
      params.delete('query');
    }

    docType === 'all' ? params.delete('doc_type') : params.set('doc_type', docType);
    status === 'all' ? params.delete('status') : params.set('status', status);
    confidence === 'all' ? params.delete('confidence') : params.set('confidence', confidence);
    rerank ? params.set('rerank', 'true') : params.set('rerank', 'false');

    router.push(`${pathname}?${params.toString()}`);
  };

  return (
    <Card className="overflow-hidden border-white/[0.06] bg-gradient-to-br from-card to-secondary/20">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <SlidersHorizontal className="h-5 w-5 text-primary" />
          Query Controls
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="grid gap-3 md:grid-cols-12 md:gap-4">
          <div className="md:col-span-5">
            <Input placeholder="Search across chunks, concepts, and citations..." value={query} onChange={(event) => setQuery(event.target.value)} />
          </div>
          <div className="md:col-span-2">
            <Select value={docType} onValueChange={setDocType}>
              <SelectTrigger>
                <SelectValue placeholder="Doc Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="policy">Policy</SelectItem>
                <SelectItem value="runbook">Runbook</SelectItem>
                <SelectItem value="spec">Spec</SelectItem>
                <SelectItem value="incident">Incident</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="md:col-span-2">
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger>
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="archived">Archived</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="md:col-span-2">
            <Select value={confidence} onValueChange={setConfidence}>
              <SelectTrigger>
                <SelectValue placeholder="Confidence" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Any Confidence</SelectItem>
                <SelectItem value="high">High</SelectItem>
                <SelectItem value="medium">Medium</SelectItem>
                <SelectItem value="low">Low</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center justify-between gap-3 md:col-span-1 md:justify-end">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Switch checked={rerank} onCheckedChange={setRerank} />
              <span>Rerank</span>
            </div>
            <Button type="submit" className="w-full md:w-auto">
              <Search className="mr-2 h-4 w-4" />
              Run
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
