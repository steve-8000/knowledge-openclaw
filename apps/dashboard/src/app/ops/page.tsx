import { JobTable } from '@/components/ops/job-table';
import { PipelineStatus } from '@/components/ops/pipeline-status';

export default function OpsPage(): JSX.Element {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="font-[var(--font-heading)] text-3xl font-semibold">Ops Monitoring</h1>
        <p className="text-muted-foreground">Track pipeline health, outbox lag, worker activity, and recent indexing jobs.</p>
      </div>
      <PipelineStatus />
      <JobTable />
    </div>
  );
}
