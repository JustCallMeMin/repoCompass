"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Shell } from "../../../../components/Shell";
import { Repository, listOrgRepositories } from "../../../../lib/api";

export default function OrgRepositoriesPage() {
  const { orgId } = useParams<{ orgId: string }>();
  const [repos, setRepos] = useState<Repository[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!orgId) return;
    listOrgRepositories(orgId)
      .then((data) => setRepos(data.data ?? []))
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [orgId]);

  return (
    <Shell>
      <section className="grid gap-5">
        <div className="border-b border-ink/10 pb-4">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Organization</p>
          <h2 className="font-display text-3xl font-semibold md:text-4xl">Repositories</h2>
        </div>

        {loading && <p className="text-sm text-ink/50">Loading…</p>}

        {error && (
          <div className="border border-rust/30 bg-rust/10 p-6 text-center">
            <p className="font-display text-xl text-rust">Could not load repositories.</p>
            <p className="mt-2 text-sm text-rust/80">{error}</p>
            <Link href="/" className="mt-4 inline-block text-sm text-moss underline hover:text-ink">
              Back to scan page
            </Link>
          </div>
        )}

        {!loading && !error && repos.length === 0 && (
          <p className="text-sm text-ink/50">No repositories in this organization yet. Run a scan to add one.</p>
        )}

        <div className="grid gap-3">
          {repos.map((r) => (
            <Link
              key={r.id}
              href={`/repositories/${r.id}`}
              id={`repo-card-${r.id}`}
              className="flex items-center justify-between border border-ink/15 bg-paper/90 p-4 hover:shadow-[4px_4px_0_rgba(23,32,29,0.1)] transition"
            >
              <div>
                <p className="font-semibold">{r.full_name || r.name}</p>
                <p className="font-mono text-xs text-ink/40">{r.id}</p>
              </div>
              <span className="border border-ink/15 px-2 py-0.5 text-xs font-bold uppercase tracking-[0.12em]">
                {r.status}
              </span>
            </Link>
          ))}
        </div>
      </section>
    </Shell>
  );
}
