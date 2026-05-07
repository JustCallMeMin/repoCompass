"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Shell } from "../../../components/Shell";
import { StatusPill } from "../../../components/StatusPill";
import { Badge, Card, EmptyState, ErrorState, LoadingState } from "../../../components/ui";
import {
  Assessment,
  FindingDetail,
  ReportMetadata,
  ScanDetail,
  getScan,
  getScanAssessment,
  listScanFindings,
  listScanReports,
} from "../../../lib/api";

export function ScanClient({ scanId, view = "overview" }: { scanId: string; view?: "overview" | "findings" | "recommendations" }) {
  const [scan, setScan] = useState<ScanDetail | null>(null);
  const [assessment, setAssessment] = useState<Assessment | null>(null);
  const [findings, setFindings] = useState<FindingDetail[]>([]);
  const [reports, setReports] = useState<ReportMetadata[]>([]);
  const [severity, setSeverity] = useState("all");
  const [category, setCategory] = useState("all");
  const [sort, setSort] = useState("severity");
  const [openFinding, setOpenFinding] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([getScan(scanId), getScanAssessment(scanId).catch(() => null), listScanFindings(scanId), listScanReports(scanId).catch(() => [])])
      .then(([nextScan, nextAssessment, nextFindings, nextReports]) => {
        setScan(nextScan);
        setAssessment(nextAssessment);
        setFindings(nextFindings);
        setReports(nextReports);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load scan"))
      .finally(() => setLoading(false));
  }, [scanId]);

  const categories = useMemo(() => Array.from(new Set(findings.map((finding) => finding.category).filter(Boolean))), [findings]);
  const filteredFindings = useMemo(() => {
    const severityRank: Record<string, number> = { high: 0, medium: 1, low: 2 };
    return findings
      .filter((finding) => severity === "all" || finding.severity === severity)
      .filter((finding) => category === "all" || finding.category === category)
      .sort((a, b) => {
        if (sort === "path") return firstPath(a).localeCompare(firstPath(b));
        return (severityRank[a.severity] ?? 9) - (severityRank[b.severity] ?? 9);
      });
  }, [findings, severity, category, sort]);
  const recommendations = filteredFindings.flatMap((finding) =>
    finding.recommendations.map((recommendation) => ({ ...recommendation, finding_id: finding.id, severity: finding.severity, category: finding.category })),
  );

  return (
    <Shell>
      <section className="grid gap-5">
        {loading ? <LoadingState label="Loading scan" /> : null}
        {error ? <ErrorState message={error} /> : null}
        {scan ? (
          <>
            <Card className="grid gap-4 md:grid-cols-[1fr_auto] md:items-start">
              <div>
                <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Scan</p>
                <h2 className="mt-3 break-all font-display text-3xl font-semibold">{scan.id}</h2>
                <div className="mt-3 flex flex-wrap gap-2">
                  <StatusPill value={scan.status} />
                  {assessment ? <Badge tone={assessment.overall_score >= 80 ? "good" : assessment.overall_score >= 60 ? "warn" : "bad"}>{assessment.label || "assessment"}</Badge> : null}
                </div>
              </div>
              <nav className="flex flex-wrap gap-2 text-xs font-bold uppercase tracking-[0.12em]">
                <Link className="border border-ink/15 px-3 py-2 hover:bg-field" href={`/scans/${scanId}`}>Overview</Link>
                <Link className="border border-ink/15 px-3 py-2 hover:bg-field" href={`/scans/${scanId}/findings`}>Findings</Link>
                <Link className="border border-ink/15 px-3 py-2 hover:bg-field" href={`/scans/${scanId}/recommendations`}>Recommendations</Link>
              </nav>
            </Card>

            {view === "overview" ? (
              <div className="grid gap-5 lg:grid-cols-[0.85fr_1.15fr]">
                <Card>
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Assessment</p>
                  {assessment ? <AssessmentPanel assessment={assessment} /> : <EmptyState title="No assessment." body="This scan has no persisted assessment yet." />}
                </Card>
                <Card>
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Reports</p>
                  {reports.length ? (
                    <div className="mt-3 grid gap-2">
                      {reports.map((report) => (
                        <div key={report.id} className="flex items-center justify-between border border-ink/10 bg-field/70 p-3 text-sm">
                          <span>{report.format}</span>
                          <span className="text-ink/50">{report.content_type}</span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <EmptyState title="No report artifacts." body="Report metadata will appear here when generated." />
                  )}
                </Card>
                <Card className="lg:col-span-2">
                  <FindingsToolbar severity={severity} category={category} sort={sort} categories={categories} onSeverity={setSeverity} onCategory={setCategory} onSort={setSort} />
                  <FindingsList findings={filteredFindings} openFinding={openFinding} setOpenFinding={setOpenFinding} />
                </Card>
              </div>
            ) : view === "findings" ? (
              <Card>
                <FindingsToolbar severity={severity} category={category} sort={sort} categories={categories} onSeverity={setSeverity} onCategory={setCategory} onSort={setSort} />
                <FindingsList findings={filteredFindings} openFinding={openFinding} setOpenFinding={setOpenFinding} />
              </Card>
            ) : (
              <Card>
                <p className="text-sm font-bold uppercase tracking-[0.18em] text-rust">Recommendations</p>
                {recommendations.length ? (
                  <div className="mt-4 grid gap-3">
                    {recommendations.map((recommendation) => (
                      <article key={`${recommendation.finding_id}-${recommendation.title}`} className="border border-ink/15 bg-white p-4">
                        <div className="flex flex-wrap gap-2">
                          <Badge tone={recommendation.priority === "high" ? "bad" : "neutral"}>{recommendation.priority}</Badge>
                          <Badge>{recommendation.category}</Badge>
                        </div>
                        <h3 className="mt-3 font-display text-2xl">{recommendation.title}</h3>
                        <p className="mt-2 text-sm text-ink/70">{recommendation.rationale}</p>
                        <p className="mt-2 text-sm font-semibold">{recommendation.action}</p>
                        <button onClick={() => setOpenFinding(recommendation.finding_id ?? null)} className="mt-3 text-sm text-moss underline">Linked finding {recommendation.finding_id}</button>
                      </article>
                    ))}
                  </div>
                ) : (
                  <EmptyState title="No recommendations." body="Healthy scans may not produce recommendations." />
                )}
              </Card>
            )}
          </>
        ) : !loading && !error ? (
          <EmptyState title="Scan not found." body="Open a scan from repository history." />
        ) : null}
      </section>
    </Shell>
  );
}

function FindingsToolbar({
  severity,
  category,
  sort,
  categories,
  onSeverity,
  onCategory,
  onSort,
}: {
  severity: string;
  category: string;
  sort: string;
  categories: string[];
  onSeverity: (value: string) => void;
  onCategory: (value: string) => void;
  onSort: (value: string) => void;
}) {
  return (
    <div className="mb-4 grid gap-3 md:grid-cols-[1fr_1fr_1fr]">
      <select value={severity} onChange={(event) => onSeverity(event.target.value)} className="border border-ink/20 bg-white px-3 py-2 text-sm">
        <option value="all">All severities</option>
        <option value="high">High</option>
        <option value="medium">Medium</option>
        <option value="low">Low</option>
      </select>
      <select value={category} onChange={(event) => onCategory(event.target.value)} className="border border-ink/20 bg-white px-3 py-2 text-sm">
        <option value="all">All categories</option>
        {categories.map((item) => <option key={item} value={item}>{item}</option>)}
      </select>
      <select value={sort} onChange={(event) => onSort(event.target.value)} className="border border-ink/20 bg-white px-3 py-2 text-sm">
        <option value="severity">Sort severity</option>
        <option value="path">Sort path</option>
      </select>
    </div>
  );
}

function FindingsList({ findings, openFinding, setOpenFinding }: { findings: FindingDetail[]; openFinding: string | null; setOpenFinding: (value: string | null) => void }) {
  if (!findings.length) {
    return <EmptyState title="No findings." body="Empty state is valid for healthy repositories." />;
  }
  return (
    <div className="grid gap-4">
      {findings.map((finding) => (
        <article key={finding.id} className="border border-ink/15 bg-white p-4">
          <button className="w-full text-left" onClick={() => setOpenFinding(openFinding === finding.id ? null : finding.id)}>
            <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
              <div>
                <div className="flex flex-wrap gap-2">
                  <Badge tone={finding.severity === "high" ? "bad" : finding.severity === "medium" ? "warn" : "neutral"}>{finding.severity}</Badge>
                  <Badge>{finding.category}</Badge>
                </div>
                <h3 className="mt-2 font-display text-2xl font-semibold">{finding.title}</h3>
              </div>
              <p className="text-sm text-ink/60">{finding.analyzer_id}</p>
            </div>
            <p className="mt-3 text-sm leading-6 text-ink/70">{finding.message}</p>
          </button>
          {openFinding === finding.id ? (
            <div className="mt-4 border-t border-ink/10 pt-4">
              <p className="text-sm font-bold uppercase tracking-[0.14em] text-moss">Evidence</p>
              <ul className="mt-2 grid gap-2 text-sm text-ink/70">
                {finding.evidence.map((item, index) => <li key={`${finding.id}-ev-${index}`}>{item.path ? `${item.path}: ` : ""}{item.message}</li>)}
              </ul>
            </div>
          ) : null}
        </article>
      ))}
    </div>
  );
}

function AssessmentPanel({ assessment }: { assessment: Assessment }) {
  return (
    <div className="mt-4 grid gap-4">
      <div className="grid grid-cols-3 gap-3">
        <MiniMetric label="Score" value={`${assessment.overall_score}`} />
        <MiniMetric label="Label" value={assessment.label || "-"} />
        <MiniMetric label="Findings" value={`${assessment.finding_count}`} />
      </div>
      <p className="text-sm text-ink/70">{assessment.overall_score >= 80 ? "Repository is in good onboarding shape." : assessment.overall_score >= 60 ? "Repository needs targeted cleanup." : "Repository has high onboarding risk."}</p>
      <Breakdown title="Severity" values={assessment.severity_counts ?? {}} />
      <Breakdown title="Category scores" values={assessment.category_scores ?? {}} />
    </div>
  );
}

function Breakdown({ title, values }: { title: string; values: Record<string, number> }) {
  const entries = Object.entries(values);
  if (!entries.length) return null;
  return (
    <div>
      <p className="text-xs font-bold uppercase tracking-[0.14em] text-ink/45">{title}</p>
      <div className="mt-2 grid gap-2">
        {entries.map(([key, value]) => <div key={key} className="flex justify-between border border-ink/10 bg-field/70 px-3 py-2 text-sm"><span>{key}</span><span>{value}</span></div>)}
      </div>
    </div>
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

function firstPath(finding: FindingDetail) {
  return finding.evidence.find((item) => item.path)?.path ?? "";
}
