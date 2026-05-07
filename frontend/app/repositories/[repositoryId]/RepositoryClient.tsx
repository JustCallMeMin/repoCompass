"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Shell } from "../../../components/Shell";
import { StatusPill } from "../../../components/StatusPill";
import { Badge, Button, Card, EmptyState, ErrorState, LoadingState } from "../../../components/ui";
import {
  FindingDetail,
  MetricPoint,
  Repository,
  ScanSummary,
  createRepositoryScan,
  getRepository,
  listRepositoryMetrics,
  listRepositoryScans,
  listScanFindings,
} from "../../../lib/api";

export function RepositoryClient({ repositoryId }: { repositoryId: string }) {
  const [repository, setRepository] = useState<Repository | null>(null);
  const [scans, setScans] = useState<ScanSummary[]>([]);
  const [metrics, setMetrics] = useState<MetricPoint[]>([]);
  const [findings, setFindings] = useState<FindingDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [error, setError] = useState("");
  const latest = scans[0];
  const highFindings = findings.filter((finding) => finding.severity === "high");
  const recommendations = findings.flatMap((finding) =>
    finding.recommendations.map((recommendation) => ({ ...recommendation, finding_id: finding.id, severity: finding.severity, category: finding.category })),
  );
  const trendDirection = useMemo(() => {
    if (metrics.length < 2) return "No trend yet";
    const first = metrics[0].value;
    const last = metrics[metrics.length - 1].value;
    if (last > first) return "Score improved";
    if (last < first) return "Score declined";
    return "Score stable";
  }, [metrics]);

  function load() {
    setLoading(true);
    setError("");
    Promise.all([getRepository(repositoryId), listRepositoryScans(repositoryId), listRepositoryMetrics(repositoryId)])
      .then(async ([repo, nextScans, nextMetrics]) => {
        setRepository(repo);
        setScans(nextScans);
        setMetrics(nextMetrics);
        if (nextScans[0]?.scan_id) {
          setFindings(await listScanFindings(nextScans[0].scan_id));
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load repository"))
      .finally(() => setLoading(false));
  }

  useEffect(load, [repositoryId]);

  async function triggerScan() {
    setScanning(true);
    setError("");
    try {
      await createRepositoryScan(repositoryId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to trigger scan");
    } finally {
      setScanning(false);
    }
  }

  return (
    <Shell>
      <section className="grid gap-5">
        {loading ? <LoadingState label="Loading repository" /> : null}
        {error ? <ErrorState message={error} /> : null}
        {repository ? (
          <>
            <Card className="grid gap-4 md:grid-cols-[1fr_auto] md:items-start">
              <div>
                <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Repository</p>
                <h2 className="mt-3 break-all font-display text-3xl font-semibold md:text-4xl">{repository.full_name || repository.name}</h2>
                <div className="mt-3 flex flex-wrap gap-2">
                  <Badge>{repository.provider}</Badge>
                  <Badge tone={repository.status === "active" ? "good" : "warn"}>{repository.status}</Badge>
                  {latest ? <StatusPill value={latest.status} /> : <Badge tone="warn">no scans</Badge>}
                </div>
                <p className="mt-3 break-all font-mono text-xs text-ink/45">{repository.url || repository.id}</p>
              </div>
              <Button onClick={triggerScan} disabled={scanning}>{scanning ? "Scanning" : "Run scan"}</Button>
            </Card>

            <div className="grid gap-5 lg:grid-cols-[0.75fr_1.25fr]">
              <div className="grid gap-5">
                <Card>
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Latest scan</p>
                  {latest ? (
                    <div className="mt-4 grid gap-3">
                      <Link href={`/scans/${latest.scan_id}`} className="break-all font-mono text-sm text-moss underline">{latest.scan_id}</Link>
                      <div className="grid grid-cols-3 gap-3">
                        <MiniMetric label="Score" value={latest.score.toString()} />
                        <MiniMetric label="Findings" value={latest.finding_count.toString()} />
                        <MiniMetric label="Label" value={latest.label || "-"} />
                      </div>
                    </div>
                  ) : (
                    <EmptyState title="No scan yet." body="Run a scan to populate repository health." />
                  )}
                </Card>
                <Card>
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Score trend</p>
                  <p className="mt-2 text-sm text-ink/60">{trendDirection}</p>
                  <ScoreTrend points={metrics.slice(-12)} />
                </Card>
                <Card>
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Insights</p>
                  <ul className="mt-3 grid gap-2 text-sm text-ink/70">
                    <li>{latest ? "Latest scan available for review." : "No recent scan. Run one before reviewing health."}</li>
                    <li>{highFindings.length ? `${highFindings.length} high severity findings need attention.` : "No high severity findings in latest scan."}</li>
                    <li>{recommendations.length ? `${recommendations.length} recommendations ready.` : "No recommendations available yet."}</li>
                  </ul>
                </Card>
              </div>

              <Card>
                <div className="mb-5 flex items-center justify-between gap-4">
                  <h3 className="font-display text-3xl font-semibold">Scan history</h3>
                  {latest ? <Link href={`/scans/${latest.scan_id}/recommendations`} className="text-sm font-bold uppercase tracking-[0.14em] text-rust">Recommendations</Link> : null}
                </div>
                <div className="divide-y divide-ink/10">
                  {scans.map((scan, index) => (
                    <Link key={scan.scan_id} href={`/scans/${scan.scan_id}`} className="grid gap-3 py-4 md:grid-cols-[1fr_auto_auto] md:items-center">
                      <div>
                        <p className="break-all font-mono text-sm">{scan.scan_id}</p>
                        <p className="mt-1 text-sm text-ink/60">Score {scan.score} / findings {scan.finding_count} {index === 0 ? "/ latest" : ""}</p>
                      </div>
                      <StatusPill value={scan.status} />
                      <span className="text-sm text-ink/60">{scan.label || "unlabeled"}</span>
                    </Link>
                  ))}
                </div>
                {!scans.length ? <EmptyState title="No scans found." body="Run a scan to build history." /> : null}
              </Card>
            </div>
          </>
        ) : !loading && !error ? (
          <EmptyState title="Repository not found." body="Return to repository list and choose an available repository." />
        ) : null}
      </section>
    </Shell>
  );
}

function MiniMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-ink/10 bg-field/70 p-3">
      <p className="text-xs uppercase tracking-[0.14em] text-ink/45">{label}</p>
      <p className="mt-1 font-display text-2xl">{value}</p>
    </div>
  );
}

function ScoreTrend({ points }: { points: MetricPoint[] }) {
  if (points.length < 2) {
    return <p className="mt-4 text-sm text-ink/60">Not enough metric points for a trend.</p>;
  }
  return (
    <div className="mt-4 flex h-32 items-end gap-2 border-l border-b border-ink/20 px-2">
      {points.map((point) => (
        <div key={`${point.scan_id}-${point.captured_at}`} className="flex flex-1 flex-col items-center gap-2">
          <div className="w-full bg-moss" style={{ height: `${Math.max(10, Math.min(100, point.value))}%` }} />
          <span className="text-[10px] text-ink/50">{Math.round(point.value)}</span>
        </div>
      ))}
    </div>
  );
}
