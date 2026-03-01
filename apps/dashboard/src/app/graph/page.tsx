import { KnowledgeGraph } from '@/components/graph/knowledge-graph';

export default function GraphPage(): JSX.Element {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Knowledge Graph</h1>
        <p className="text-muted-foreground">Explore document relations with interactive ego graphs and relation filters.</p>
      </div>
      <KnowledgeGraph />
    </div>
  );
}
