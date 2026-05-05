"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { Shell } from "../components/Shell";
import { StatusPill } from "../components/StatusPill";
import { ScanResponse, createScan } from "../lib/api";

export default function DashboardPage() {
  const [sourceType, setSourceType] = useState<"local" | "github">("local");
  const [value, setValue] = useState("./testdata/fixtures/local-repositories/good-onboarding-repo");
  const [result, setResult] = useState<ScanResponse | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      const nextResult = await createScan(
        sourceType === "local"
          ? { source_type: "local", path: value }
          : { source_type: "github", url: value },
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
      <section className="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <form onSubmit={submit} className="border border-ink/15 bg-paper/90 p-5 shadow-[10px_10px_0_rgba(23,32,29,0.09)]">
          <div className="mb-8 flex flex-col gap-3">
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">New scan</p>
            <h2 className="font-display text-3xl font-semibold md:text-5xl">Point RepoCompass at a codebase.</h2>
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
          <input
            value={value}
            onChange={(event) => setValue(event.target.value)}
            className="mb-5 w-full border border-ink/20 bg-white px-4 py-3 text-sm outline-none focus:border-rust"
          />
          <button
            disabled={loading}
            className="w-full bg-rust px-5 py-3 text-sm font-bold uppercase tracking-[0.18em] text-white transition hover:bg-ink disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading ? "Scanning" : "Run scan"}
          </button>
          {error ? <p className="mt-4 border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p> : null}
        </form>

        <div className="border border-ink/15 bg-ink p-5 text-paper shadow-[10px_10px_0_rgba(178,91,53,0.2)]">
          <p className="mb-3 text-sm font-bold uppercase tracking-[0.18em] text-gold">Latest result</p>
          {result ? (
            <div className="flex h-full flex-col justify-between gap-8">
              <div>
                <StatusPill value={result.status} />
                <div className="mt-6 grid grid-cols-3 gap-3">
                  <Metric label="Score" value={`${result.assessment_score}`} />
                  <Metric label="Findings" value={`${result.finding_count}`} />
                  <Metric label="Analyzers" value={`${result.analyzers_processed}`} />
                </div>
              </div>
              <div className="grid gap-3 text-sm text-paper/70">
                <CodeLine label="Scan" value={result.scan_id} />
                <CodeLine label="Repository" value={result.repository_id} />
                <CodeLine label="Snapshot" value={result.snapshot_id} />
              </div>
              <div className="grid gap-3 md:grid-cols-2">
                <Link className="border border-paper/20 px-4 py-3 text-center text-sm font-bold uppercase tracking-[0.14em] hover:bg-paper hover:text-ink" href={`/repositories/${result.repository_id}`}>
                  Open repository
                </Link>
                <Link className="border border-paper/20 px-4 py-3 text-center text-sm font-bold uppercase tracking-[0.14em] hover:bg-paper hover:text-ink" href={`/scans/${result.scan_id}`}>
                  Open findings
                </Link>
              </div>
            </div>
          ) : (
            <div className="flex min-h-80 items-end">
              <p className="max-w-md font-display text-3xl leading-tight text-paper/80">
                No scan selected. Run a local or GitHub scan to populate this panel.
              </p>
            </div>
          )}
        </div>
      </section>
    </Shell>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-paper/15 p-4">
      <p className="text-xs uppercase tracking-[0.16em] text-paper/50">{label}</p>
      <p className="mt-2 font-display text-4xl">{value}</p>
    </div>
  );
}

function CodeLine({ label, value }: { label: string; value: string }) {
  return (
    <p>
      <span className="text-paper/45">{label}: </span>
      <span className="break-all font-mono text-paper">{value}</span>
    </p>
  );
}
