import { Suspense } from 'react';
import { SearchForm } from '@/components/search/search-form';
import { SearchResults } from '@/components/search/search-results';

export default function SearchPage(): JSX.Element {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Semantic Search</h1>
        <p className="text-muted-foreground">Query the indexed knowledge base with BM25 + ANN ranking, reranking, and citations.</p>
      </div>
      <Suspense fallback={<div className="text-muted-foreground">Loading...</div>}>
        <SearchForm />
        <SearchResults />
      </Suspense>
    </div>
  );
}
