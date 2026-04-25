"use client";

import { useEffect, useRef, useState } from "react";
import { Trash2, Upload, X } from "lucide-react";

import { api, type UploadFileResult } from "@/lib/api";
import { cn } from "@/lib/cn";

const ACCEPTED = ".xlsx,.docx,.txt,.md,.png,.jpg,.jpeg,.gif,.webp,.pdf";
const MAX_FILES = 10;
const MAX_BYTES = 20 * 1024 * 1024;

function formatSize(bytes: number): string {
  return bytes < 1024 * 1024
    ? `${(bytes / 1024).toFixed(1)} KB`
    : `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function validate(files: File[]): string | null {
  if (files.length > MAX_FILES) return `Maximum ${MAX_FILES} files per upload.`;
  const oversized = files.find((f) => f.size > MAX_BYTES);
  if (oversized) return `"${oversized.name}" exceeds the 20 MB limit.`;
  return null;
}

export function UploadModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [staged, setStaged] = useState<File[]>([]);
  const [dragOver, setDragOver] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [results, setResults] = useState<UploadFileResult[] | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  function addFiles(incoming: FileList | null) {
    if (!incoming) return;
    const merged = [...staged, ...Array.from(incoming)];
    const err = validate(merged);
    setValidationError(err);
    setStaged(merged);
  }

  function removeFile(index: number) {
    const next = staged.filter((_, i) => i !== index);
    setValidationError(validate(next));
    setStaged(next);
  }

  async function handleUpload() {
    const err = validate(staged);
    if (err) { setValidationError(err); return; }
    setUploading(true);
    setUploadError(null);
    try {
      const { results: res } = await api.uploadFiles(staged);
      setResults(res);
      onSuccess();
    } catch (e) {
      setUploadError(e instanceof Error ? e.message : "Upload failed.");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="upload-modal-title"
        className="w-full max-w-xl rounded-lg border border-neutral-800 bg-neutral-900 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-neutral-800 px-5 py-4">
          <h2 id="upload-modal-title" className="text-base font-semibold">Upload files</h2>
          <button
            type="button"
            aria-label="Close upload dialog"
            onClick={onClose}
            className="text-neutral-500 hover:text-neutral-200"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-4 p-5">
          {/* Dropzone */}
          {!results && (
            <div
              className={cn(
                "flex cursor-pointer flex-col items-center justify-center gap-2 rounded border-2 border-dashed px-4 py-8 text-sm transition-colors",
                dragOver
                  ? "border-blue-500 bg-blue-900/20 text-blue-300"
                  : "border-neutral-700 text-neutral-500 hover:border-neutral-500 hover:text-neutral-300",
              )}
              onClick={() => inputRef.current?.click()}
              onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
              onDragLeave={() => setDragOver(false)}
              onDrop={(e) => { e.preventDefault(); setDragOver(false); addFiles(e.dataTransfer.files); }}
            >
              <Upload className="h-6 w-6" />
              <span>Click or drag files here</span>
              <span className="text-xs text-neutral-600">
                Excel, Word, Text, Markdown, Image, PDF — max {MAX_FILES} files, 20 MB each
              </span>
            </div>
          )}

          <input
            ref={inputRef}
            type="file"
            multiple
            accept={ACCEPTED}
            className="hidden"
            onChange={(e) => addFiles(e.target.files)}
          />

          {/* Validation error */}
          {validationError && (
            <p className="rounded border border-red-800 bg-red-900/20 px-3 py-2 text-sm text-red-400">
              {validationError}
            </p>
          )}

          {/* Upload error */}
          {uploadError && (
            <p className="rounded border border-red-800 bg-red-900/20 px-3 py-2 text-sm text-red-400">
              {uploadError}
            </p>
          )}

          {/* Staged file list */}
          {!results && staged.length > 0 && (
            <ul className="space-y-1">
              {staged.map((f, i) => (
                <li key={i} className="flex items-center justify-between rounded border border-neutral-800 bg-neutral-950 px-3 py-2 text-sm">
                  <span className="truncate text-neutral-200">{f.name}</span>
                  <div className="ml-3 flex shrink-0 items-center gap-3">
                    <span className="text-xs text-neutral-500">{formatSize(f.size)}</span>
                    <button
                      onClick={() => removeFile(i)}
                      className="text-neutral-500 hover:text-red-400"
                      aria-label={`Remove ${f.name}`}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}

          {/* Results */}
          {results && (
            <ul className="space-y-1">
              {results.map((r, i) => (
                <li key={i} className="flex items-center justify-between rounded border border-neutral-800 bg-neutral-950 px-3 py-2 text-sm">
                  <span className="truncate text-neutral-200">{r.filename}</span>
                  <div className="ml-3 flex shrink-0 items-center gap-2 text-xs">
                    {r.status === "ok" ? (
                      <>
                        <span className="text-neutral-500">{r.chunks_created} chunk{r.chunks_created !== 1 ? "s" : ""}</span>
                        {r.replaced && <span className="text-amber-400">replaced</span>}
                        <span className="rounded bg-green-900 px-1.5 py-0.5 text-green-300">ok</span>
                      </>
                    ) : (
                      <span className="rounded bg-red-900 px-1.5 py-0.5 text-red-300" title={r.error_message ?? undefined}>
                        error
                      </span>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 border-t border-neutral-800 px-5 py-4">
          {results ? (
            <button
              onClick={() => { onSuccess(); onClose(); }}
              className="rounded bg-blue-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-blue-500"
            >
              Done
            </button>
          ) : (
            <>
              <button
                onClick={onClose}
                className="rounded border border-neutral-700 px-4 py-1.5 text-sm text-neutral-400 hover:border-neutral-500 hover:text-neutral-200"
              >
                Cancel
              </button>
              <button
                onClick={handleUpload}
                disabled={staged.length === 0 || !!validationError || uploading}
                className="flex items-center gap-2 rounded bg-blue-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
              >
                {uploading && <span className="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />}
                {uploading ? "Uploading…" : `Upload ${staged.length > 0 ? staged.length : ""} file${staged.length !== 1 ? "s" : ""}`}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
