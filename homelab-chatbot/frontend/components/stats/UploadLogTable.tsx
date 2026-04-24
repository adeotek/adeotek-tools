import type { UploadLogEntry } from "@/lib/api";

function formatSize(bytes: number): string {
  return bytes < 1024 * 1024
    ? `${(bytes / 1024).toFixed(1)} KB`
    : `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function UploadLogTable({ entries }: { entries: UploadLogEntry[] }) {
  if (entries.length === 0) {
    return <p className="text-sm text-neutral-500">No files uploaded yet.</p>;
  }

  return (
    <div className="overflow-hidden rounded border border-neutral-800">
      <table className="w-full text-sm">
        <thead className="bg-neutral-900 text-left text-xs text-neutral-500">
          <tr>
            <th className="px-4 py-2">Filename</th>
            <th className="px-4 py-2">Type</th>
            <th className="px-4 py-2 text-right">Size</th>
            <th className="px-4 py-2">Date</th>
            <th className="px-4 py-2 text-center">Status</th>
            <th className="px-4 py-2 text-right">Chunks</th>
            <th className="px-4 py-2 text-center">Replaced</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-neutral-800">
          {entries.map((e) => (
            <tr key={e.id} className="bg-neutral-900/50">
              <td className="max-w-[180px] truncate px-4 py-2 font-mono text-xs" title={e.filename}>
                {e.filename}
              </td>
              <td className="px-4 py-2 text-neutral-400">{e.mime_type}</td>
              <td className="px-4 py-2 text-right text-neutral-400">{formatSize(e.file_size)}</td>
              <td className="px-4 py-2 text-neutral-400">
                {new Date(e.uploaded_at).toLocaleString()}
              </td>
              <td className="px-4 py-2 text-center">
                {e.status === "ok" ? (
                  <span className="rounded bg-green-900 px-1.5 py-0.5 text-xs font-medium text-green-300">
                    ok
                  </span>
                ) : (
                  <span
                    className="rounded bg-red-900 px-1.5 py-0.5 text-xs font-medium text-red-300"
                    title={e.error_message ?? undefined}
                  >
                    error
                  </span>
                )}
              </td>
              <td className="px-4 py-2 text-right text-neutral-400">{e.chunks_created}</td>
              <td className="px-4 py-2 text-center text-neutral-400">
                {e.replaced ? (
                  <span className="rounded bg-amber-900/40 px-1.5 py-0.5 text-xs text-amber-400">yes</span>
                ) : (
                  "—"
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
