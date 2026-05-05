"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Shell } from "../../../components/Shell";
import { StatusPill } from "../../../components/StatusPill";
import { MetricPoint, ScanSummary, listRepositoryMetrics, listRepositoryScans } from "../../../lib/api";

export function RepositoryClient({ repositoryId }: { repositoryId: string }) {
  const [scans, setScans] = useState<ScanSummary[]>([]);
  const [metrics, setMetrics] = useState<MetricPoint[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([listRepositoryScans(repositoryId), listRepositoryMetrics(repositoryId)])
      .then(([nextScans, nextMetrics]) => {
        setScans(nextScans);
        setMetrics(nextMetrics);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load repository"));
  }, [repositoryId]);

  return (
    <Shell>
      <section className="grid gap-5 lg:grid-cols-[0.75fr_1.25fr]">
        <aside className="border border-ink/15 bg-paper/90 p-5">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Repository</p>
          <h2 className="mt-4 break-all font-display text-3xl font-semibold">{repositoryId}</h2>
          <div className="mt-8 border-t border-ink/10 pt-5">
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Metric trend</p>
            {metrics.length ? (
              <div className="mt-4 flex items-end gap-2">
                {metrics.slice(-12).map((point) => (
                  <div key={point.scan_id} className="flex flex-1 flex-col items-center gap-2">
                    <div className="w-full bg-moss" style={{ height: `${Math.max(12, point.value)}px` }} />
                    <span className="text-xs text-ink/60">{point.value}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="mt-4 text-sm text-ink/60">No metric points yet.</p>
            )}
          </div>
        </aside>
        <div className="border border-ink/15 bg-paper/90 p-5">
          <div className="mb-5 flex items-center justify-between gap-4">
            <h3 className="font-display text-3xl font-semibold">Scan history</h3>
            <Link href="/" className="text-sm font-bold uppercase tracking-[0.14em] text-rust">
              New scan
            </Link>
          </div>
          {error ? <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p> : null}
          <div className="divide-y divide-ink/10">
            {scans.map((scan) => (
              <Link
                key={scan.scan_id}
                href={`/scans/${scan.scan_id}`}
                className="grid gap-3 py-4 md:grid-cols-[1fr_auto_auto] md:items-center"
              >
                <div>
                  <p className="break-all font-mono text-sm">{scan.scan_id}</p>
                  <p className="mt-1 text-sm text-ink/60">
                    Score {scan.score} / findings {scan.finding_count}
                  </p>
                </div>
                <StatusPill value={scan.status} />
                <span className="text-sm text-ink/60">{scan.label || "unlabeled"}</span>
              </Link>
            ))}
          </div>
          {!scans.length && !error ? <p className="py-10 text-sm text-ink/60">No scans found for this repository.</p> : null}
        </div>
      </section>
    </Shell>
  );
}
