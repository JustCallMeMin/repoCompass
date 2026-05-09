"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Shell } from "../../components/Shell";
import { Badge, Card, EmptyState, ErrorState, LoadingState } from "../../components/ui";
import { Repository, listRepositories } from "../../lib/api";

export default function RepositoriesPage() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [query, setQuery] = useState("");
  const [provider, setProvider] = useState("all");
  const [status, setStatus] = useState("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listRepositories()
      .then(setRepositories)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load repositories"))
      .finally(() => setLoading(false));
  }, []);

  const providers = useMemo(() => Array.from(new Set(repositories.map((repo) => repo.provider).filter(Boolean))), [repositories]);
  const statuses = useMemo(() => Array.from(new Set(repositories.map((repo) => repo.status).filter(Boolean))), [repositories]);
  const filtered = repositories.filter((repo) => {
    const haystack = `${repo.name} ${repo.full_name}`.toLowerCase();
    return (
      haystack.includes(query.toLowerCase()) &&
      (provider === "all" || repo.provider === provider) &&
      (status === "all" || repo.status === status)
    );
  });

  return (
    <Shell>
      <section className="grid gap-5">
        <div className="flex flex-col gap-3 border-b border-ink/10 pb-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Repositories</p>
            <h2 className="font-display text-3xl font-semibold md:text-4xl">Repository inventory</h2>
          </div>
          <Link href="/dashboard" className="text-sm font-bold uppercase tracking-[0.14em] text-rust">New scan</Link>
        </div>

        <Card className="grid gap-3 md:grid-cols-[1fr_180px_180px]">
          <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search repositories" className="border border-ink/20 bg-white px-3 py-2 text-sm outline-none focus:border-rust" />
          <select value={provider} onChange={(event) => setProvider(event.target.value)} className="border border-ink/20 bg-white px-3 py-2 text-sm">
            <option value="all">All providers</option>
            {providers.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select value={status} onChange={(event) => setStatus(event.target.value)} className="border border-ink/20 bg-white px-3 py-2 text-sm">
            <option value="all">All statuses</option>
            {statuses.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
        </Card>

        {loading ? <LoadingState label="Loading repositories" /> : null}
        {error ? <ErrorState message={error} /> : null}
        {!loading && !error && filtered.length === 0 ? <EmptyState title="No repositories found." body="Run a scan or adjust the filters." /> : null}

        <div className="grid gap-3">
          {filtered.map((repo) => (
            <Link key={repo.id} href={`/repositories/${repo.id}`} className="grid gap-3 border border-ink/15 bg-paper/90 p-4 hover:shadow-[4px_4px_0_rgba(23,32,29,0.1)] md:grid-cols-[1fr_auto_auto] md:items-center">
              <div>
                <p className="font-semibold">{repo.full_name || repo.name}</p>
                <p className="font-mono text-xs text-ink/40">{repo.id}</p>
              </div>
              <Badge>{repo.provider || "unknown"}</Badge>
              <Badge tone={repo.status === "active" ? "good" : "warn"}>{repo.status || "unknown"}</Badge>
            </Link>
          ))}
        </div>
      </section>
    </Shell>
  );
}
