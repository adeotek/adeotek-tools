"use client";

import { useCallback, useEffect, useState } from "react";

import { ChatThread } from "@/components/chat/ChatThread";
import { Composer } from "@/components/chat/Composer";
import { api, type Conversation, type Message } from "@/lib/api";
import { parseSseStream, type DisplayMessage } from "@/lib/stream";

export default function Home() {
  const [provider, setProvider] = useState("anthropic");
  const [model, setModel] = useState("claude-sonnet-4-6");
  const [convId, setConvId] = useState<string | null>(null);
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [streaming, setStreaming] = useState(false);

  useEffect(() => {
    api.getSettings().then((s) => {
      setProvider(s.default_provider);
      setModel(s.default_model);
    }).catch(() => {});
  }, []);

  const send = useCallback(
    async (text: string) => {
      setMessages((prev) => [
        ...prev,
        { id: `u-${Date.now()}`, role: "user", content: text },
        { id: `a-${Date.now()}`, role: "assistant", content: "", partial: true },
      ]);
      setStreaming(true);
      try {
        const resp = await fetch("/api/chat", {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            conv_id: convId,
            message: text,
            provider,
            model,
          }),
        });
        if (!resp.ok) throw new Error("chat failed");

        for await (const ev of parseSseStream(resp)) {
          if (ev.kind === "conversation") {
            setConvId(ev.id);
          } else if (ev.kind === "text-delta") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = {
                  ...last,
                  content: last.content + ev.text,
                };
              }
              return copy;
            });
          } else if (ev.kind === "tool-call" || ev.kind === "tool-result") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                const events = [...(last.toolEvents ?? []),
                  { kind: ev.kind, name: ev.name, summary: (ev as {summary?: string}).summary ?? "" }];
                copy[copy.length - 1] = { ...last, toolEvents: events };
              }
              return copy;
            });
          } else if (ev.kind === "done") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = { ...last, partial: false };
              }
              return copy;
            });
          }
        }
      } finally {
        setStreaming(false);
      }
    },
    [convId, provider, model],
  );

  return (
    <main className="flex h-screen flex-col">
      <header className="flex items-center gap-2 border-b border-neutral-800 p-3 text-sm">
        <span className="text-neutral-400">Provider:</span>
        <span>{provider}</span>
        <span className="text-neutral-400">· Model:</span>
        <span>{model}</span>
      </header>
      <ChatThread messages={messages} />
      <Composer onSend={send} disabled={streaming} />
    </main>
  );
}
