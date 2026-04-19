"use client";

import useSWR from "swr";

import { api, type Settings } from "@/lib/api";

const fetcher = () => api.getSettings();

export function ProviderPicker({
  provider,
  model,
  onChange,
}: {
  provider: string;
  model: string;
  onChange: (p: string, m: string) => void;
}) {
  const { data } = useSWR<Settings>("/api/settings", fetcher);
  if (!data) {
    return <span className="text-xs text-neutral-500">loading providers…</span>;
  }

  const current = data.providers.find((p) => p.id === provider);
  const models = current?.models ?? [];

  return (
    <div className="flex items-center gap-2 text-sm">
      <select
        aria-label="provider"
        className="rounded border border-neutral-700 bg-neutral-950 px-2 py-1"
        value={provider}
        onChange={(e) => {
          const next = data.providers.find((p) => p.id === e.target.value);
          const firstModel = next?.models[0] ?? "";
          onChange(e.target.value, firstModel);
        }}
      >
        {data.providers.map((p) => (
          <option key={p.id} value={p.id} disabled={!p.available}>
            {p.id}
            {!p.available ? " (unavailable)" : ""}
          </option>
        ))}
      </select>
      <select
        aria-label="model"
        className="rounded border border-neutral-700 bg-neutral-950 px-2 py-1"
        value={model}
        onChange={(e) => onChange(provider, e.target.value)}
      >
        {models.map((m) => (
          <option key={m} value={m}>
            {m}
          </option>
        ))}
      </select>
    </div>
  );
}
