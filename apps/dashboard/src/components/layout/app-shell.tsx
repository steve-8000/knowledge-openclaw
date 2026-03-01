'use client';

import { useEffect, useState } from 'react';

import { Header } from '@/components/layout/header';
import { Sidebar } from '@/components/layout/sidebar';

const TENANT_STORAGE_KEY = 'ki-db:tenant-id';
const DEFAULT_TENANT_ID = '00000000-0000-0000-0000-000000000001';

type AppShellProps = {
  children: React.ReactNode;
};

export function AppShell({ children }: AppShellProps): JSX.Element {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [tenantId, setTenantId] = useState(DEFAULT_TENANT_ID);

  useEffect(() => {
    const stored = window.localStorage.getItem(TENANT_STORAGE_KEY);
    if (stored) {
      setTenantId(stored);
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(TENANT_STORAGE_KEY, tenantId);
  }, [tenantId]);

  return (
    <div className="flex min-h-screen">
      <Sidebar mobileOpen={mobileOpen} onClose={() => setMobileOpen(false)} />
      <div className="flex min-h-screen flex-1 flex-col lg:pl-0">
        <Header tenantId={tenantId} onTenantChange={setTenantId} onOpenSidebar={() => setMobileOpen(true)} />
        <main className="mx-auto w-full max-w-[1600px] flex-1 p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}
