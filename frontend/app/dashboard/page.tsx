"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { Shell } from "../../components/Shell";
import { StatusPill } from "../../components/StatusPill";
import { Card, EmptyState, ErrorState, LoadingState } from "../../components/ui";
import { Repository, ScanResponse, createScan, listRepositories } from "../../lib/api";

export default function DashboardPage() {
  const [sourceType, setSourceType] = useState<"local" | "github">("local");
  const [value, setValue] = useState("./testdata/fixtures/local-repositories/good-onboarding-repo");
  const [result, setResult] = useState<ScanResponse | null>(null);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingRepos, setLoadingRepos] = useState(true);
  const providerCount = useMemo(() => new Set(repositories.map((repo) => repo.provider)).size, [repositories]);

  useEffect(() => {
    listRepositories()
      .then(setRepositories)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load repositories"))
      .finally(() => setLoadingRepos(false));
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      const nextResult = await createScan(
        sourceType === "local" ? { source_type: "local", path: value } : { source_type: "github", url: value },
      );
      setResult(nextResult);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Scan failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Shell>
      <section className="grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
        <form onSubmit={submit} className="border border-ink/15 bg-paper/90 p-5 shadow-[10px_10px_0_rgba(23,32,29,0.09)]">
          <div className="mb-8 flex flex-col gap-3">
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Dashboard</p>
            <h2 className="font-display text-3xl font-semibold md:text-5xl">Understand repository onboarding health.</h2>
          </div>
          <div className="mb-4 grid grid-cols-2 border border-ink/20">
            {(["local", "github"] as const).map((item) => (
              <button
                key={item}
                type="button"
                onClick={() => {
                  setSourceType(item);
                  setValue(item === "local" ? "./testdata/fixtures/local-repositories/good-onboarding-repo" : "https://github.com/spf13/cobra");
                }}
                className={`px-4 py-3 text-sm font-bold uppercase tracking-[0.16em] ${
                  sourceType === item ? "bg-ink text-paper" : "bg-paper text-ink hover:bg-field"
                }`}
              >
                {item}
              </button>
            ))}
          </div>
          <label className="mb-2 block text-sm font-bold text-ink/75">
            {sourceType === "local" ? "Local repository path" : "Public GitHub repository URL"}
          </label>
          <input value={value} onChange={(event) => setValue(event.target.value)} className="mb-5 w-full border border-ink/20 bg-white px-4 py-3 text-sm outline-none focus:border-rust" />
          <button disabled={loading} className="w-full bg-rust px-5 py-3 text-sm font-bold uppercase tracking-[0.18em] text-white transition hover:bg-ink disabled:cursor-not-allowed disabled:opacity-60">
            {loading ? "Scanning" : "Run scan"}
          </button>
          {error ? <ErrorState message={error} /> : null}
        </form>

        <div className="grid gap-5">
          <div className="grid gap-3 md:grid-cols-3">
            <Card><Metric label="Repositories" value={repositories.length.toString()} /></Card>
            <Card><Metric label="Providers" value={providerCount.toString()} /></Card>
            <Card><Metric label="Latest score" value={result ? result.assessment_score.toString() : "-"} /></Card>
          </div>
          <Card className="bg-ink text-paper">
            <p className="mb-3 text-sm font-bold uppercase tracking-[0.18em] text-gold">Latest result</p>
            {result ? (
              <div className="grid gap-5">
                <StatusPill value={result.status} />
                <div className="grid grid-cols-3 gap-3">
                  <Metric label="Score" value={`${result.assessment_score}`} />
                  <Metric label="Findings" value={`${result.finding_count}`} />
                  <Metric label="Analyzers" value={`${result.analyzers_processed}`} />
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <Link className="border border-paper/20 px-4 py-3 text-center text-sm font-bold uppercase tracking-[0.14em] hover:bg-paper hover:text-ink" href={`/repositories/${result.repository_id}`}>
                    Repository
                  </Link>
                  <Link className="border border-paper/20 px-4 py-3 text-center text-sm font-bold uppercase tracking-[0.14em] hover:bg-paper hover:text-ink" href={`/scans/${result.scan_id}`}>
                    Scan detail
                  </Link>
                </div>
              </div>
            ) : loadingRepos ? (
              <LoadingState label="Loading dashboard data" />
            ) : (
              <EmptyState title="No scan selected." body="Run a local or GitHub scan, or open an existing repository." />
            )}
          </Card>
        </div>
      </section>
    </Shell>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs uppercase tracking-[0.16em] text-current/50">{label}</p>
      <p className="mt-2 font-display text-4xl">{value}</p>
    </div>
  );
}
