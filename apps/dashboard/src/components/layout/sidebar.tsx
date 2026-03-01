'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Activity, BrainCircuit, Network, Search, ShieldCheck, X } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

type SidebarProps = {
  mobileOpen: boolean;
  onClose: () => void;
};

const navItems = [
  { href: '/search', label: 'Search', icon: Search },
  { href: '/graph', label: 'Graph', icon: Network },
  { href: '/curation', label: 'Curation', icon: BrainCircuit },
  { href: '/quality', label: 'Quality', icon: ShieldCheck },
  { href: '/ops', label: 'Ops', icon: Activity }
];

export function Sidebar({ mobileOpen, onClose }: SidebarProps): JSX.Element {
  const pathname = usePathname();

  return (
    <>
      <aside className={cn('fixed inset-y-0 left-0 z-40 w-72 transform border-r border-white/[0.06] bg-[hsl(var(--sidebar-background))] text-[hsl(var(--sidebar-foreground))] transition-transform duration-300 ease-out lg:relative lg:translate-x-0', mobileOpen ? 'translate-x-0' : '-translate-x-full')}>
        <div className="flex h-16 items-center justify-between border-b border-white/[0.06] px-4">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.3em] text-primary/60">ki-db</p>
            <h1 className="text-lg font-semibold">Knowledge Index</h1>
          </div>
          <Button variant="ghost" size="icon" className="text-white hover:bg-white/10 lg:hidden" onClick={onClose} aria-label="Close navigation">
            <X className="h-4 w-4" />
          </Button>
        </div>
        <nav className="space-y-1 p-3">
          {navItems.map((item, idx) => {
            const Icon = item.icon;
            const active = pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={onClose}
                className={cn(
                  'flex animate-fade-in-up items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-200',
                  active ? 'bg-white/[0.06] text-white shadow-[inset_0_0_0_1px_rgba(255,255,255,0.08)]' : 'text-white/50 hover:bg-white/[0.04] hover:text-white/80'
                )}
                style={{ animationDelay: `${idx * 70}ms` }}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>
      </aside>
      {mobileOpen ? (
        <button
          type="button"
          className="fixed inset-0 z-30 bg-black/60 backdrop-blur-sm lg:hidden"
          onClick={onClose}
          aria-label="Close menu overlay"
        />
      ) : null}
    </>
  );
}
