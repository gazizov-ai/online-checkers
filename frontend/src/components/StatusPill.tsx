type StatusTone = "neutral" | "good" | "warn" | "danger";

export function StatusPill({
  tone = "neutral",
  children,
}: {
  tone?: StatusTone;
  children: string;
}) {
  return <span className={`status-pill status-pill--${tone}`}>{children}</span>;
}
