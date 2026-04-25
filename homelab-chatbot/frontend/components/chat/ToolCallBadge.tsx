import { cn } from "@/lib/cn";

export function ToolCallBadge({ name }: { name: string }) {
  const icon = name.includes("inventory") ? "📊" : "🔍";
  const label = name.includes("inventory") ? "querying inventory" : "searching docs";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border border-neutral-700",
        "bg-neutral-800 px-2 py-0.5 text-xs text-neutral-300",
      )}
    >
      <span>{icon}</span>
      <span>{label}</span>
    </span>
  );
}
