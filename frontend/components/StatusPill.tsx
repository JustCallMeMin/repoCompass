export function StatusPill({ value }: { value: string }) {
  const tone =
    value === "completed"
      ? "border-moss/30 bg-moss/10 text-moss"
      : value === "failed"
        ? "border-rust/30 bg-rust/10 text-rust"
        : "border-gold/40 bg-gold/15 text-ink";

  return (
    <span className={`inline-flex w-fit items-center border px-2.5 py-1 text-xs font-bold uppercase tracking-[0.18em] ${tone}`}>
      {value}
    </span>
  );
}
