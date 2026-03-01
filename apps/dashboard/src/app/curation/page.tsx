import { CurationQueue } from '@/components/curation/curation-queue';

export default function CurationPage(): JSX.Element {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Curation Inbox</h1>
        <p className="text-muted-foreground">Phase 2 scaffold for human-in-the-loop approval and rejection workflows.</p>
      </div>
      <CurationQueue />
    </div>
  );
}
