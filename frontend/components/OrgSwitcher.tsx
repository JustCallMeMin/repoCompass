"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { Organization, listOrganizations } from "../lib/api";

const PERSONAL_ORG_ID = "00000000-0000-0000-0000-000000000000";
const PERSONAL_ORG: Organization = {
  id: PERSONAL_ORG_ID,
  name: "Personal",
  created_at: "",
  updated_at: "",
};

export function OrgSwitcher() {
  const [orgs, setOrgs] = useState<Organization[]>([PERSONAL_ORG]);
  const [selected, setSelected] = useState<Organization>(PERSONAL_ORG);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    listOrganizations()
      .then((res) => {
        const list = [PERSONAL_ORG, ...(res.data ?? [])];
        setOrgs(list);
      })
      .catch(() => {/* API unavailable — show only personal org */});
  }, []);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <div className="relative" ref={ref}>
      <button
        id="org-switcher-btn"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2 border border-ink/20 bg-paper px-3 py-1.5 text-xs font-bold uppercase tracking-[0.14em] hover:border-ink/50"
      >
        <span className="max-w-[140px] truncate">{selected.name}</span>
        <span className="text-ink/40">▾</span>
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-1 min-w-[200px] border border-ink/20 bg-paper shadow-lg">
          {orgs.map((o) => (
            <button
              key={o.id}
              id={`org-opt-${o.id}`}
              className={`block w-full px-4 py-2.5 text-left text-xs font-medium hover:bg-field ${
                selected.id === o.id ? "bg-field font-bold" : ""
              }`}
              onClick={() => {
                setSelected(o);
                setOpen(false);
              }}
            >
              {o.name}
            </button>
          ))}
          <div className="border-t border-ink/10 px-4 py-2">
            <Link
              href="/organizations"
              className="text-xs text-moss underline hover:text-ink"
              onClick={() => setOpen(false)}
            >
              Manage organizations →
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
