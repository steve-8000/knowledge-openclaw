import { QualityDashboard } from '@/components/quality/quality-dashboard';

export default function QualityPage(): JSX.Element {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Quality Overview</h1>
        <p className="text-muted-foreground">Phase 2 scaffold for duplicate detection, stale content audits, and metadata completeness.</p>
      </div>
      <QualityDashboard />
    </div>
  );
}
