"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Shell } from "../../../../components/Shell";
import {
  Membership,
  Policy,
  canManageOrganization,
  currentUserId,
  listMembers,
  listPolicies,
  savePolicy,
} from "../../../../lib/api";

export default function OrgPoliciesPage() {
  const { orgId } = useParams<{ orgId: string }>();
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<string | null>(null);
  const [draftRules, setDraftRules] = useState("");
  const [saving, setSaving] = useState(false);
  const [currentRole, setCurrentRole] = useState<Membership["role"] | undefined>();
  const canManage = canManageOrganization(currentRole);

  function load() {
    if (!orgId) return;
    Promise.all([listPolicies(orgId), listMembers(orgId)])
      .then(([policyRes, memberRes]) => {
        setPolicies(policyRes.data ?? []);
        setCurrentRole(memberRes.data?.find((m) => m.user_id === currentUserId)?.role);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }

  useEffect(load, [orgId]);

  async function handleSave(policy: Policy) {
    if (!orgId || !canManage) return;
    setSaving(true);
    try {
      const rules = JSON.parse(draftRules) as Record<string, unknown>;
      await savePolicy(orgId, { id: policy.id, name: policy.name, rules });
      setEditing(null);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Shell>
      <section className="grid gap-5">
        <div className="border-b border-ink/10 pb-4">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Organization</p>
          <h2 className="font-display text-3xl font-semibold md:text-4xl">Policies</h2>
          <p className="mt-1 text-sm text-ink/50">
            Assessment policies define minimum score thresholds and required files for this organization.
          </p>
          <p className="mt-2 text-xs font-bold uppercase tracking-[0.14em] text-ink/40">
            Current role: {currentRole ?? "none"} {canManage ? "· manage enabled" : "· read-only"}
          </p>
        </div>

        {loading && <p className="text-sm text-ink/50">Loading…</p>}
        {error && <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p>}

        {!loading && policies.length === 0 && (
          <p className="text-sm text-ink/50">No policies configured for this organization.</p>
        )}

        <div className="grid gap-4">
          {policies.map((p) => (
            <div
              key={p.id}
              id={`policy-card-${p.id}`}
              className="border border-ink/15 bg-paper/90 p-5"
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="font-semibold">{p.name}</p>
                  <p className="mt-0.5 font-mono text-xs text-ink/40">{p.id}</p>
                </div>
                <button
                  id={`policy-edit-${p.id}`}
                  onClick={() => {
                    if (!canManage) return;
                    setEditing(p.id);
                    setDraftRules(JSON.stringify(p.rules, null, 2));
                  }}
                  disabled={!canManage}
                  className="border border-ink/20 px-3 py-1 text-xs font-bold uppercase tracking-[0.12em] hover:bg-field disabled:cursor-not-allowed disabled:opacity-40"
                >
                  Edit
                </button>
              </div>

              {editing === p.id ? (
                <div className="mt-4">
                  <label className="mb-1 block text-xs font-bold text-ink/60">Rules (JSON)</label>
                  <textarea
                    id={`policy-rules-editor-${p.id}`}
                    value={draftRules}
                    onChange={(e) => setDraftRules(e.target.value)}
                    rows={6}
                    className="w-full border border-ink/20 bg-white px-3 py-2 font-mono text-xs outline-none focus:border-rust"
                  />
                  <div className="mt-3 flex gap-3">
                    <button
                      id={`policy-save-${p.id}`}
                      onClick={() => handleSave(p)}
                      disabled={saving || !canManage}
                      className="bg-rust px-4 py-2 text-xs font-bold uppercase tracking-[0.14em] text-white hover:bg-ink disabled:opacity-60"
                    >
                      {saving ? "Saving…" : "Save"}
                    </button>
                    <button
                      id={`policy-cancel-${p.id}`}
                      onClick={() => setEditing(null)}
                      className="border border-ink/20 px-4 py-2 text-xs font-bold uppercase tracking-[0.14em] hover:bg-field"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <pre className="mt-3 overflow-auto rounded-sm border border-ink/10 bg-field px-3 py-2 text-xs text-ink/70">
                  {JSON.stringify(p.rules, null, 2)}
                </pre>
              )}
            </div>
          ))}
        </div>
      </section>
    </Shell>
  );
}
