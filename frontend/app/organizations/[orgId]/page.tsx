"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Shell } from "../../../components/Shell";
import {
  Membership,
  OrgInsights,
  Organization,
  getOrganization,
  getOrgInsights,
  listMembers,
  listNotifications,
} from "../../../lib/api";

export default function OrgOverviewPage() {
  const { orgId } = useParams<{ orgId: string }>();
  const [org, setOrg] = useState<Organization | null>(null);
  const [insights, setInsights] = useState<OrgInsights | null>(null);
  const [members, setMembers] = useState<Membership[]>([]);
  const [notifications, setNotifications] = useState<Array<{ id: string; title: string; severity: string; message: string }>>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!orgId) return;
    Promise.all([
      getOrganization(orgId),
      getOrgInsights(orgId).catch(() => null),
      listMembers(orgId).catch(() => ({ data: [] as Membership[] })),
      listNotifications(orgId).catch(() => []),
    ])
      .then(([orgRes, insRes, memRes, notificationRes]) => {
        setOrg(orgRes);
        setInsights(insRes ?? null);
        setMembers(Array.isArray(memRes) ? memRes : []);
        setNotifications(Array.isArray(notificationRes) ? notificationRes : []);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [orgId]);

  return (
    <Shell>
      {loading && <p className="text-sm text-ink/50">Loading…</p>}
      {error && <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p>}

      {org && (
        <section className="grid gap-6">
          <div className="border-b border-ink/10 pb-4">
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Organization</p>
            <h2 className="font-display text-3xl font-semibold md:text-4xl">{org.name}</h2>
            <p className="mt-1 font-mono text-xs text-ink/40">{org.id}</p>
          </div>

          {/* Benchmark / Insights panel — T6-022, T6-023 */}
          {insights && (
            <div className="grid gap-4 md:grid-cols-3">
              <StatCard label="Repositories" value={insights.total_repositories} />
              <StatCard label="Scans" value={insights.total_scans} />
              <StatCard label="Avg Score" value={insights.average_score} suffix="/100" />
              <StatCard label="High Risk" value={insights.high_risk_count ?? 0} />
              <StatCard label="Stale Scans" value={insights.stale_scan_count ?? 0} />
            </div>
          )}

          {insights?.insights && insights.insights.length > 0 && (
            <div className="grid gap-3">
              <p className="text-sm font-bold uppercase tracking-[0.16em] text-ink/50">Insights</p>
              {insights.insights.map((item) => (
                <div key={`${item.title}-${item.repository_id ?? item.policy_id ?? ""}`} className="border border-ink/15 bg-paper/90 p-4">
                  <p className="text-xs font-bold uppercase tracking-[0.14em] text-rust">{item.severity}</p>
                  <p className="mt-1 font-semibold">{item.title}</p>
                  <p className="mt-1 text-sm text-ink/60">{item.explanation}</p>
                  <p className="mt-2 text-sm font-medium">{item.next_action}</p>
                </div>
              ))}
            </div>
          )}

          {/* Navigation links */}
          <nav className="grid gap-3 md:grid-cols-3">
            <NavCard
              id="org-nav-repos"
              href={`/organizations/${orgId}/repositories`}
              label="Repositories"
              description="Repositories in this organization"
            />
            <NavCard
              id="org-nav-policies"
              href={`/organizations/${orgId}/policies`}
              label="Policies"
              description="Assessment policy rules"
            />
            <NavCard
              id="org-nav-members"
              href={`/organizations/${orgId}/members`}
              label="Members"
              description="Manage org membership"
            />
          </nav>

          {/* Members preview */}
          {members.length > 0 && (
            <div className="border border-ink/15 bg-paper/90 p-5">
              <p className="mb-3 text-sm font-bold uppercase tracking-[0.16em] text-ink/50">
                Members ({members.length})
              </p>
              <ul className="grid gap-2">
                {members.slice(0, 5).map((m) => (
                  <li key={m.user_id} className="flex items-center justify-between text-sm">
                    <span className="font-mono">{m.user_id}</span>
                    <span className="rounded-sm border border-ink/15 px-2 py-0.5 text-xs font-bold uppercase tracking-[0.12em]">
                      {m.role}
                    </span>
                  </li>
                ))}
              </ul>
              {members.length > 5 && (
                <Link
                  href={`/organizations/${orgId}/members`}
                  className="mt-3 block text-xs text-moss underline hover:text-ink"
                >
                  View all {members.length} members →
                </Link>
              )}
            </div>
          )}

          {notifications.length > 0 && (
            <div className="border border-ink/15 bg-paper/90 p-5">
              <p className="mb-3 text-sm font-bold uppercase tracking-[0.16em] text-ink/50">Activity</p>
              <ul className="grid gap-3">
                {notifications.slice(0, 5).map((item) => (
                  <li key={item.id} className="border-b border-ink/10 pb-2 last:border-b-0">
                    <p className="text-xs font-bold uppercase tracking-[0.12em] text-rust">{item.severity}</p>
                    <p className="font-semibold">{item.title}</p>
                    <p className="text-sm text-ink/60">{item.message}</p>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </section>
      )}
    </Shell>
  );
}

function StatCard({ label, value, suffix = "" }: { label: string; value: number; suffix?: string }) {
  return (
    <div className="border border-ink/15 bg-paper/90 p-5">
      <p className="text-xs font-bold uppercase tracking-[0.16em] text-ink/50">{label}</p>
      <p className="mt-2 font-display text-4xl font-semibold">
        {value}
        {suffix && <span className="ml-1 text-xl text-ink/40">{suffix}</span>}
      </p>
    </div>
  );
}

function NavCard({
  id,
  href,
  label,
  description,
}: {
  id: string;
  href: string;
  label: string;
  description: string;
}) {
  return (
    <Link
      id={id}
      href={href}
      className="border border-ink/15 bg-paper/90 p-5 hover:shadow-[6px_6px_0_rgba(23,32,29,0.1)] transition"
    >
      <p className="font-semibold">{label}</p>
      <p className="mt-1 text-sm text-ink/50">{description}</p>
    </Link>
  );
}
