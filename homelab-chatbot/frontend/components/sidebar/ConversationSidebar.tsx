"use client";

import Link from "next/link";
import useSWR from "swr";

import { NewChatButton } from "./NewChatButton";
import { api, type Conversation } from "@/lib/api";
import { cn } from "@/lib/cn";

const fetcher = () => api.listConversations();

export function ConversationSidebar({
  activeId,
  onSelect,
  onNew,
}: {
  activeId: string | null;
  onSelect: (c: Conversation) => void;
  onNew: () => void;
}) {
  const { data, mutate } = useSWR<Conversation[]>("/api/conv", fetcher, {
    refreshInterval: 0,
    revalidateOnFocus: true,
  });

  async function handleDelete(id: string) {
    if (!confirm("Delete this conversation?")) return;
    await api.deleteConversation(id);
    await mutate();
    if (activeId === id) onNew();
  }

  return (
    <aside className="flex w-64 flex-col border-r border-neutral-800 bg-neutral-900 p-2">
      <NewChatButton onClick={onNew} />
      <ul className="flex-1 space-y-1 overflow-y-auto">
        {(data ?? []).map((c) => (
          <li key={c.id}>
            <div
              className={cn(
                "flex w-full items-center justify-between rounded px-2 py-1.5 text-sm",
                activeId === c.id ? "bg-neutral-800" : "hover:bg-neutral-800/60",
              )}
            >
              <button
                type="button"
                onClick={() => onSelect(c)}
                className="min-w-0 flex-1 text-left"
              >
                <span className="block truncate">{c.title}</span>
              </button>
              <button
                type="button"
                aria-label={`Delete ${c.title}`}
                onClick={() => handleDelete(c.id)}
                className="ml-2 shrink-0 text-neutral-500 hover:text-red-400"
              >
                ×
              </button>
            </div>
          </li>
        ))}
        {(data ?? []).length === 0 && (
          <li className="p-2 text-center text-sm text-neutral-500">
            No conversations yet.
          </li>
        )}
      </ul>
      <nav className="mt-2 border-t border-neutral-800 pt-2">
        <Link href="/stats" className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-sm font-semibold text-neutral-300 hover:bg-neutral-800/60 hover:text-white">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          Data ingestion
        </Link>
        <Link href="/settings" className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-sm font-semibold text-neutral-300 hover:bg-neutral-800/60 hover:text-white">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Settings
        </Link>
      </nav>
    </aside>
  );
}
