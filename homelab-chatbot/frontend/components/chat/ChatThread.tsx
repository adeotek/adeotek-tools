"use client";

import { useEffect, useRef } from "react";

import { StreamingMessage } from "./StreamingMessage";
import type { DisplayMessage } from "@/lib/stream";

export function ChatThread({ messages }: { messages: DisplayMessage[] }) {
  const endRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div className="flex-1 space-y-3 overflow-y-auto p-4">
      {messages.length === 0 && (
        <p className="mt-8 text-center text-neutral-500">
          Start a conversation about your home lab.
        </p>
      )}
      {messages.map((m) => (
        <StreamingMessage key={m.id} message={m} />
      ))}
      <div ref={endRef} />
    </div>
  );
}
