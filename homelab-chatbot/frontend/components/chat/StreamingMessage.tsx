"use client";

import ReactMarkdown from "react-markdown";

import { ToolCallBadge } from "./ToolCallBadge";
import type { DisplayMessage } from "@/lib/stream";

export function StreamingMessage({ message }: { message: DisplayMessage }) {
  const isUser = message.role === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-3xl rounded-lg px-4 py-3 ${
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
        <div className="prose prose-invert max-w-none text-sm">
          <ReactMarkdown>{message.content || "…"}</ReactMarkdown>
        </div>
      </div>
    </div>
  );
}
