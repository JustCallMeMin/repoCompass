import Link from "next/link";
import type { ReactNode } from "react";
import { AuthGate } from "./AuthGate";
import { OrgSwitcher } from "./OrgSwitcher";

export function Shell({ children, protectedRoute = true }: { children: ReactNode; protectedRoute?: boolean }) {
  return (
    <main className="grain min-h-screen px-5 py-5 text-ink md:px-8">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <header className="flex flex-col gap-4 border-b border-ink/15 pb-5 md:flex-row md:items-end md:justify-between">
          <Link href="/" className="group">
            <p className="text-xs font-bold uppercase tracking-[0.24em] text-rust">RepoCompass</p>
            <h1 className="font-display text-4xl font-semibold leading-none md:text-6xl">
              Scan control room
            </h1>
          </Link>
          <div className="flex flex-col items-start gap-3 md:items-end">
            <div className="flex items-center gap-3">
              <OrgSwitcher />
              <Link
                href="/organizations"
                id="nav-organizations"
                className="text-xs font-bold uppercase tracking-[0.14em] text-ink/60 hover:text-ink"
              >
                Organizations
              </Link>
            </div>
            <p className="max-w-xl text-sm leading-6 text-ink/70">
              Trigger scans, inspect persisted history, and review findings without leaving the product surface.
            </p>
          </div>
        </header>
        {protectedRoute ? <AuthGate>{children}</AuthGate> : children}
      </div>
    </main>
  );
}
