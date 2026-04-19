"use client";

export function NewChatButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="mb-2 w-full rounded bg-blue-600 p-2 text-sm font-medium"
    >
      + New chat
    </button>
  );
}
