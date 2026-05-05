"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Shell } from "../../../components/Shell";
import { FindingDetail, listScanFindings } from "../../../lib/api";

export function ScanClient({ scanId }: { scanId: string }) {
  const [findings, setFindings] = useState<FindingDetail[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    listScanFindings(scanId)
      .then(setFindings)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load findings"));
  }, [scanId]);

  return (
    <Shell>
      <section className="border border-ink/15 bg-paper/90 p-5">
        <div className="mb-6 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Scan findings</p>
            <h2 className="mt-3 break-all font-display text-3xl font-semibold">{scanId}</h2>
          </div>
          <Link href="/" className="text-sm font-bold uppercase tracking-[0.14em] text-rust">
            Back to dashboard
          </Link>
        </div>
        {error ? <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p> : null}
        <div className="grid gap-4">
          {findings.map((finding) => (
            <article key={finding.id} className="border border-ink/15 bg-white p-4">
              <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                <div>
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-rust">{finding.severity}</p>
                  <h3 className="mt-2 font-display text-2xl font-semibold">{finding.title}</h3>
                </div>
                <p className="text-sm text-ink/60">{finding.analyzer_id}</p>
              </div>
              <p className="mt-3 text-sm leading-6 text-ink/70">{finding.message}</p>
              {finding.recommendations?.length ? (
                <div className="mt-4 border-t border-ink/10 pt-4">
                  <p className="text-sm font-bold uppercase tracking-[0.14em] text-moss">Recommendations</p>
                  <ul className="mt-2 grid gap-2 text-sm text-ink/70">
                    {finding.recommendations.map((item) => (
                      <li key={`${finding.id}-${item.title}`}>
                        {item.title}: {item.action}
                      </li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </article>
          ))}
        </div>
        {!findings.length && !error ? (
          <div className="border border-dashed border-ink/25 bg-field/70 p-10 text-center">
            <p className="font-display text-3xl">No findings for this scan.</p>
            <p className="mt-2 text-sm text-ink/60">Empty state is valid for healthy repositories.</p>
          </div>
        ) : null}
      </section>
    </Shell>
  );
}
