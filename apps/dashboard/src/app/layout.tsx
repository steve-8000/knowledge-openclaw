import type { Metadata } from 'next';
import { Fraunces, IBM_Plex_Mono } from 'next/font/google';

import '@/app/globals.css';
import { AppShell } from '@/components/layout/app-shell';

const fraunces = Fraunces({
  subsets: ['latin'],
  variable: '--font-heading',
  weight: ['500', '600', '700']
});

const plexMono = IBM_Plex_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  weight: ['400', '500']
});

export const metadata: Metadata = {
  title: 'ki-db Dashboard',
  description: 'Admin interface for search, graph, curation, quality, and ops monitoring.'
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>): JSX.Element {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className={`${fraunces.variable} ${plexMono.variable} font-sans`}>
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}
