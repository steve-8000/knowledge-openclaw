'use client';

import { Menu, Building2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

type HeaderProps = {
  tenantId: string;
  onTenantChange: (value: string) => void;
  onOpenSidebar: () => void;
};

const tenantOptions = [
  { label: 'Primary Tenant', value: '00000000-0000-0000-0000-000000000001' },
  { label: 'Sandbox Tenant', value: '00000000-0000-0000-0000-000000000002' }
];

export function Header({ tenantId, onTenantChange, onOpenSidebar }: HeaderProps): JSX.Element {
  return (
    <header className="sticky top-0 z-20 border-b bg-background/70 backdrop-blur-xl">
      <div className="flex h-16 items-center justify-between gap-3 px-4 md:px-6">
        <div className="flex items-center gap-3">
          <Button variant="outline" size="icon" className="lg:hidden" onClick={onOpenSidebar} aria-label="Open navigation">
            <Menu className="h-4 w-4" />
          </Button>
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-muted-foreground">Admin</p>
            <h2 className="text-base font-semibold md:text-lg">Knowledge Index Dashboard</h2>
          </div>
        </div>
        <div className="flex items-center gap-2 md:gap-3">
          <Building2 className="h-4 w-4 text-muted-foreground" />
          <Select value={tenantId} onValueChange={onTenantChange}>
            <SelectTrigger className="w-[220px] bg-card">
              <SelectValue placeholder="Select tenant" />
            </SelectTrigger>
            <SelectContent>
              {tenantOptions.map((tenant) => (
                <SelectItem key={tenant.value} value={tenant.value}>
                  {tenant.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </header>
  );
}
