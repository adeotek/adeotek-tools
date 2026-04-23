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

  async function handleDelete(id: string, e: React.MouseEvent) {
    e.stopPropagation();
    if (!confirm("Delete this conversation?")) return;
    await api.deleteConversation(id);
    await mutate();
    if (activeId === id) onNew();
  }

  return (
    <aside className="flex w-64 flex-col border-r border-neutral-800 bg-neutral-900 p-2">
      <nav className="mb-2 border-b border-neutral-800 pb-2">
        <Link href="/stats" className="flex w-full rounded px-2 py-1.5 text-sm font-semibold text-neutral-300 hover:bg-neutral-800/60 hover:text-white">
          Statistics
        </Link>
        <Link href="/settings" className="flex w-full rounded px-2 py-1.5 text-sm font-semibold text-neutral-300 hover:bg-neutral-800/60 hover:text-white">
          Settings
        </Link>
      </nav>
      <NewChatButton onClick={onNew} />
      <ul className="flex-1 space-y-1 overflow-y-auto">
        {(data ?? []).map((c) => (
          <li key={c.id}>
            <button
              onClick={() => onSelect(c)}
              className={cn(
                "flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-sm",
                activeId === c.id ? "bg-neutral-800" : "hover:bg-neutral-800/60",
              )}
            >
              <span className="truncate">{c.title}</span>
              <span
                role="button"
                aria-label={`delete ${c.title}`}
                onClick={(e) => handleDelete(c.id, e)}
                className="ml-2 text-neutral-500 hover:text-red-400"
              >
                ×
              </span>
            </button>
          </li>
        ))}
        {(data ?? []).length === 0 && (
          <li className="p-2 text-center text-sm text-neutral-500">
            No conversations yet.
          </li>
        )}
      </ul>
    </aside>
  );
}
