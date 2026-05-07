"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Shell } from "../../../../components/Shell";
import { Membership, addMember, canManageOrganization, currentUserId, listMembers } from "../../../../lib/api";

export default function OrgMembersPage() {
  const { orgId } = useParams<{ orgId: string }>();
  const [members, setMembers] = useState<Membership[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [newUserId, setNewUserId] = useState("");
  const [newRole, setNewRole] = useState<Membership["role"]>("member");
  const [adding, setAdding] = useState(false);
  const currentRole = members.find((m) => m.user_id === currentUserId)?.role;
  const canManage = canManageOrganization(currentRole);

  function load() {
    if (!orgId) return;
    listMembers(orgId)
      .then((res) => setMembers(res ?? []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }

  useEffect(load, [orgId]);

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!newUserId.trim() || !orgId || !canManage) return;
    setAdding(true);
    try {
      await addMember(orgId, newUserId.trim(), newRole);
      setNewUserId("");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add member");
    } finally {
      setAdding(false);
    }
  }

  return (
    <Shell>
      <section className="grid gap-5">
        <div className="border-b border-ink/10 pb-4">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-moss">Organization</p>
          <h2 className="font-display text-3xl font-semibold md:text-4xl">Members</h2>
          <p className="mt-2 text-xs font-bold uppercase tracking-[0.14em] text-ink/40">
            Current role: {currentRole ?? "none"} {canManage ? "· manage enabled" : "· read-only"}
          </p>
        </div>

        {loading && <p className="text-sm text-ink/50">Loading…</p>}
        {error && <p className="border border-rust/30 bg-rust/10 p-3 text-sm text-rust">{error}</p>}

        <div className="grid gap-6 md:grid-cols-[1fr_340px]">
          {/* Member list */}
          <div className="border border-ink/15 bg-paper/90 p-5">
            {members.length === 0 && !loading ? (
              <p className="text-sm text-ink/50">No members yet.</p>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-ink/10 text-left text-xs font-bold uppercase tracking-[0.14em] text-ink/40">
                    <th className="pb-2">User</th>
                    <th className="pb-2">Role</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-ink/5">
                  {members.map((m) => (
                    <tr key={m.user_id} id={`member-row-${m.user_id}`}>
                      <td className="py-2 font-mono">{m.user_id}</td>
                      <td className="py-2">
                        <span className="rounded-sm border border-ink/15 px-2 py-0.5 text-xs font-bold uppercase tracking-[0.12em]">
                          {m.role}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {/* Add member form */}
          <form
            id="add-member-form"
            onSubmit={handleAdd}
            className="border border-ink/15 bg-paper/90 p-5 h-fit"
          >
            <p className="mb-4 text-sm font-bold uppercase tracking-[0.16em] text-ink/50">
              Add member
            </p>
            <label className="mb-1 block text-xs font-bold text-ink/60">User ID</label>
            <input
              id="add-member-user-id"
              value={newUserId}
              onChange={(e) => setNewUserId(e.target.value)}
              disabled={!canManage}
              className="mb-3 w-full border border-ink/20 bg-white px-3 py-2 text-sm outline-none focus:border-rust"
              placeholder="user_abc"
            />
            <label className="mb-1 block text-xs font-bold text-ink/60">Role</label>
            <select
              id="add-member-role"
              value={newRole}
              onChange={(e) => setNewRole(e.target.value as Membership["role"])}
              disabled={!canManage}
              className="mb-4 w-full border border-ink/20 bg-white px-3 py-2 text-sm outline-none"
            >
              {(["owner", "admin", "member", "viewer"] as const).map((r) => (
                <option key={r} value={r}>{r}</option>
              ))}
            </select>
            <button
              id="add-member-submit"
              type="submit"
              disabled={adding || !canManage}
              className="w-full bg-rust px-4 py-2.5 text-xs font-bold uppercase tracking-[0.16em] text-white hover:bg-ink disabled:opacity-60"
            >
              {canManage ? (adding ? "Adding…" : "Add member") : "Owner/admin only"}
            </button>
          </form>
        </div>
      </section>
    </Shell>
  );
}
