import Link from "next/link";
import type { ReactNode } from "react";

export function Card({ children, className = "" }: { children: ReactNode; className?: string }) {
  return <div className={`border border-ink/15 bg-paper/90 p-5 ${className}`}>{children}</div>;
}

export function Button({
  children,
  disabled,
  onClick,
  type = "button",
}: {
  children: ReactNode;
  disabled?: boolean;
  onClick?: () => void;
  type?: "button" | "submit";
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className="bg-rust px-4 py-2.5 text-xs font-bold uppercase tracking-[0.14em] text-white hover:bg-ink disabled:cursor-not-allowed disabled:opacity-60"
    >
      {children}
    </button>
  );
}

export function Badge({ children, tone = "neutral" }: { children: ReactNode; tone?: "good" | "warn" | "bad" | "neutral" }) {
  const classes = {
    good: "border-moss/30 bg-moss/10 text-moss",
    warn: "border-gold/50 bg-gold/15 text-ink",
    bad: "border-rust/30 bg-rust/10 text-rust",
    neutral: "border-ink/15 bg-field text-ink/70",
  };
  return <span className={`inline-flex border px-2 py-0.5 text-xs font-bold uppercase tracking-[0.12em] ${classes[tone]}`}>{children}</span>;
}

export function LoadingState({ label = "Loading" }: { label?: string }) {
  return <p className="border border-ink/10 bg-field/70 p-5 text-sm text-ink/60">{label}...</p>;
}

export function EmptyState({ title, body }: { title: string; body: string }) {
  return (
    <div className="border border-dashed border-ink/25 bg-field/70 p-8 text-center">
      <p className="font-display text-2xl">{title}</p>
      <p className="mt-2 text-sm text-ink/60">{body}</p>
    </div>
  );
}

export function ErrorState({ message }: { message: string }) {
  return <p className="border border-rust/30 bg-rust/10 p-4 text-sm text-rust">{message}</p>;
}

export function UnauthorizedState() {
  return (
    <Card className="text-center">
      <p className="font-display text-3xl">Session required.</p>
      <p className="mt-2 text-sm text-ink/60">Start the API with dev header auth or sign in with GitHub OAuth.</p>
      <Link href="/dashboard" className="mt-4 inline-block text-sm font-bold uppercase tracking-[0.14em] text-rust">
        Back to dashboard
      </Link>
    </Card>
  );
}
