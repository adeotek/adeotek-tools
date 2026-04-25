"use client";

import Link from "next/link";
import { Upload } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { api, type Stats } from "@/lib/api";
import { UploadModal } from "@/components/stats/UploadModal";
import { UploadLogTable } from "@/components/stats/UploadLogTable";

const fetcher = () => api.getStats();

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded border border-neutral-800 bg-neutral-900 px-4 py-3">
      <p className="text-xs text-neutral-500">{label}</p>
      <p className="mt-1 text-2xl font-semibold">{value}</p>
    </div>
  );
}

function Spinner() {
  return (
    <span className="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
  );
}

export default function StatsPage() {
  const [syncError, setSyncError] = useState<string | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);

  const { data, error, mutate, isValidating } = useSWR<Stats>("/api/stats", fetcher, {
    revalidateOnFocus: false,
    refreshInterval: (d) => (d?.is_syncing ? 2000 : 0),
  });

  const isSyncing = data?.is_syncing ?? false;

  async function handleSync() {
    setSyncError(null);
    try {
      await api.triggerSync();
      await mutate();
    } catch {
      setSyncError("Failed to start sync. Check that repos are configured.");
    }
  }

  return (
    <main className="min-h-screen p-6">
      <div className="mx-auto max-w-4xl">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/" className="text-sm text-neutral-500 hover:text-neutral-300">
              &larr; Back to chat
            </Link>
            <h1 className="text-xl font-semibold">Data ingestion</h1>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => mutate()}
              disabled={isValidating || isSyncing}
              className="rounded border border-neutral-700 px-3 py-1 text-sm text-neutral-400 hover:border-neutral-500 hover:text-neutral-200 disabled:opacity-50"
            >
              {isValidating ? "Refreshing…" : "Refresh"}
            </button>
            <button
              onClick={() => setShowUploadModal(true)}
              className="flex items-center gap-1.5 rounded border border-neutral-700 px-3 py-1 text-sm text-neutral-400 hover:border-neutral-500 hover:text-neutral-200"
            >
              <Upload className="h-3.5 w-3.5" />
              Upload
            </button>
            <button
              onClick={handleSync}
              disabled={isSyncing}
              className="flex items-center gap-2 rounded bg-blue-600 px-3 py-1 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-60"
            >
              {isSyncing && <Spinner />}
              {isSyncing ? "Syncing…" : "Sync now"}
            </button>
          </div>
        </div>

        {isSyncing && (
          <div className="mb-4 flex items-center gap-2 rounded border border-blue-800 bg-blue-900/20 p-3 text-sm text-blue-300">
            <Spinner />
            Ingestion in progress — cloning repos and indexing documents…
          </div>
        )}

        {syncError && (
          <p className="mb-4 rounded border border-red-800 bg-red-900/20 p-3 text-sm text-red-400">
            {syncError}
          </p>
        )}

        {error && (
          <p className="mb-4 rounded border border-red-800 bg-red-900/20 p-3 text-sm text-red-400">
            Failed to load statistics.
          </p>
        )}

        {!data && !error && (
          <p className="text-neutral-500">Loading…</p>
        )}

        {data && (
          <div className="space-y-8">
            {/* Summary cards */}
            <section>
              <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-neutral-500">
                Overview
              </h2>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                <StatCard label="Configured repos" value={data.repos.length} />
                <StatCard label="Total chunks" value={data.total_chunks.toLocaleString()} />
                <StatCard label="Uploaded files" value={data.uploaded_files} />
                <StatCard label="Uploaded chunks" value={data.uploaded_chunks.toLocaleString()} />
              </div>
              <p className="mt-2 text-xs text-neutral-600">
                Last sync:{" "}
                {data.last_sync_at
                  ? new Date(data.last_sync_at).toLocaleString()
                  : "never"}
              </p>
            </section>

            {/* Repos */}
            <section>
              <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-neutral-500">
                Repositories
              </h2>
              {data.repos.length === 0 ? (
                <p className="text-sm text-neutral-500">
                  No repositories configured. Add repos to{" "}
                  <code className="text-neutral-400">config.yaml</code> to start ingesting docs.
                </p>
              ) : (
                <div className="space-y-3">
                  {data.repos.map((repo) => (
                    <div
                      key={repo.name}
                      className="rounded border border-neutral-800 bg-neutral-900 p-4"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div className="min-w-0">
                          <p className="font-medium">{repo.name}</p>
                          <p className="mt-0.5 truncate text-xs text-neutral-500">
                            {repo.url} &middot; branch: {repo.branch}
                          </p>
                          <p className="mt-0.5 text-xs text-neutral-600">
                            Globs: {repo.include_globs.join(", ")}
                          </p>
                        </div>
                        <div className="flex shrink-0 items-center gap-4 text-right">
                          <div>
                            <p className="text-lg font-semibold">{repo.files}</p>
                            <p className="text-xs text-neutral-500">files</p>
                          </div>
                          <div>
                            <p className="text-lg font-semibold">{repo.chunks.toLocaleString()}</p>
                            <p className="text-xs text-neutral-500">chunks</p>
                          </div>
                          <div>
                            {isSyncing ? (
                              <span className="flex items-center gap-1 rounded bg-blue-900/40 px-1.5 py-0.5 text-xs text-blue-400">
                                <Spinner />
                                syncing
                              </span>
                            ) : (
                              <span
                                className={`rounded px-1.5 py-0.5 text-xs font-medium ${
                                  repo.chunks > 0
                                    ? "bg-green-900 text-green-300"
                                    : "bg-neutral-800 text-neutral-500"
                                }`}
                              >
                                {repo.chunks > 0 ? "indexed" : "not indexed"}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            {/* KB tables */}
            {data.kb_tables.length > 0 && (
              <section>
                <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-neutral-500">
                  Inventory Tables (KB DB)
                </h2>
                <div className="overflow-hidden rounded border border-neutral-800">
                  <table className="w-full text-sm">
                    <thead className="bg-neutral-900 text-left text-xs text-neutral-500">
                      <tr>
                        <th className="px-4 py-2">Table</th>
                        <th className="px-4 py-2 text-right">Rows</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-800">
                      {data.kb_tables.map((t) => (
                        <tr key={t.table} className="bg-neutral-900/50">
                          <td className="px-4 py-2 font-mono">{t.table}</td>
                          <td className="px-4 py-2 text-right text-neutral-400">
                            {t.rows.toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            )}

            {/* Uploaded files */}
            {(data.uploaded_files > 0 || data.upload_log.length > 0) && (
              <section>
                <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-neutral-500">
                  Uploaded Files
                </h2>
                <UploadLogTable entries={data.upload_log} />
              </section>
            )}
          </div>
        )}
      </div>

      {showUploadModal && (
        <UploadModal
          onClose={() => setShowUploadModal(false)}
          onSuccess={() => mutate()}
        />
      )}
    </main>
  );
}
