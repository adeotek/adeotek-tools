"use client";

import ReactMarkdown from "react-markdown";

import { ToolCallBadge } from "./ToolCallBadge";
import type { DisplayMessage, Source } from "@/lib/stream";

function parseReActContent(content: string): { reasoning: string | null; answer: string } {
  const idx = content.lastIndexOf("Answer:");
  if (idx === -1) return { reasoning: null, answer: content };
  const reasoning = content.slice(0, idx).trim();
  const answer = content.slice(idx + "Answer:".length).trim();
  return { reasoning: reasoning || null, answer };
}

function Sources({ sources }: { sources: Source[] }) {
  return (
    <details className="mt-2">
      <summary className="cursor-pointer select-none text-xs text-neutral-500 hover:text-neutral-400">
        Sources ({sources.length})
      </summary>
      <ul className="mt-1 space-y-1">
        {sources.map((s, i) => (
          <li key={i} className="text-xs">
            <span className="text-neutral-400">{s.repo}/</span>
            <span className="text-neutral-300">{s.file_path}</span>
            {s.heading_path && (
              <span className="ml-1 text-neutral-600">— {s.heading_path}</span>
            )}
          </li>
        ))}
      </ul>
    </details>
  );
}

export function StreamingMessage({ message }: { message: DisplayMessage }) {
  const isUser = message.role === "user";
  const { reasoning, answer } = parseReActContent(message.content || "…");
  const sources = message.sources ?? [];

  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`min-w-0 max-w-3xl rounded-lg px-4 py-3 ${
          isUser ? "bg-blue-600 text-white" : "bg-neutral-800 text-neutral-100"
        }`}
      >
        {message.toolEvents?.length ? (
          <div className="mb-2 flex flex-wrap gap-1">
            {message.toolEvents.map((t, i) => (
              <ToolCallBadge key={i} name={t.name} />
            ))}
          </div>
        ) : null}
        {reasoning && (
          <details className="mb-2">
            <summary className="cursor-pointer select-none text-xs text-neutral-500 hover:text-neutral-400">
              Reasoning
            </summary>
            <div className="prose prose-invert max-w-none break-words text-xs opacity-60 mt-1">
              <ReactMarkdown>{reasoning}</ReactMarkdown>
            </div>
          </details>
        )}
        <div className="prose prose-invert max-w-none break-words text-sm">
          <ReactMarkdown>{answer || "…"}</ReactMarkdown>
        </div>
        {sources.length > 0 && <Sources sources={sources} />}
      </div>
    </div>
  );
}
