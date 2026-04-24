"use client";

import Link from "next/link";
import useSWR from "swr";

import { api, type Settings } from "@/lib/api";

const fetcher = () => api.getSettings();

export default function SettingsPage() {
  const { data, error } = useSWR<Settings>("/api/settings", fetcher);

  if (error) return <p className="p-4 text-red-400">Failed to load settings.</p>;
  if (!data) return <p className="p-4 text-neutral-500">Loading…</p>;

  return (
    <main className="p-6">
      <div className="mb-4 flex items-center gap-4">
        <Link href="/" className="text-sm text-neutral-500 hover:text-neutral-300">
          &larr; Back to chat
        </Link>
        <h1 className="text-xl font-semibold">Settings</h1>
      </div>
      <p className="mb-2 text-sm text-neutral-400">
        Default: <span className="text-neutral-200">{data.default_provider}</span> / {data.default_model}
      </p>
      <div className="space-y-4">
        {data.providers.map((p) => (
          <section key={p.id} className="rounded border border-neutral-800 p-4">
            <h2 className="mb-2 font-medium">
              {p.id}{" "}
              {p.available ? (
                <span className="text-green-400 text-xs">available</span>
              ) : (
                <span className="text-red-400 text-xs">unavailable</span>
              )}
            </h2>
            <p className="text-sm text-neutral-400">
              {p.models.length === 0
                ? "No models available."
                : `Models: ${p.models.join(", ")}`}
            </p>
            {p.id === "ollama" && p.models.length > 0 && (
              <p className="mt-2 text-xs text-neutral-500">
                Tool-capable: {p.tool_capable.join(", ") || "none"}
              </p>
            )}
          </section>
        ))}
      </div>
    </main>
  );
}
