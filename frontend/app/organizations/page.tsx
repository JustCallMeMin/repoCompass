"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Shell } from "../../components/Shell";
import { Organization, listOrganizations } from "../../lib/api";

export default function OrganizationsPage() {
  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listOrganizations()
      .then((res) => setOrgs(res ?? []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <Shell>
      <section className="grid gap-5">
        <div className="border-b border-ink/10 pb-4">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Organizations</p>
          <h2 className="font-display text-3xl font-semibold md:text-4xl">All organizations</h2>
        </div>

        {loading && <p className="text-sm text-ink/50">Loading…</p>}
        {error && (
          <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p>
        )}

        {!loading && !error && orgs.length === 0 && (
          <p className="text-sm text-ink/50">No organizations found.</p>
        )}

        <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          {orgs.map((org) => (
            <Link
              key={org.id}
              id={`org-card-${org.id}`}
              href={`/organizations/${org.id}`}
              className="border border-ink/15 bg-paper/90 p-5 shadow-[6px_6px_0_rgba(23,32,29,0.07)] hover:shadow-[6px_6px_0_rgba(23,32,29,0.15)] transition"
            >
              <p className="font-display text-xl font-semibold">{org.name}</p>
              <p className="mt-1 font-mono text-xs text-ink/40 truncate">{org.id}</p>
            </Link>
          ))}
        </div>
      </section>
    </Shell>
  );
}
