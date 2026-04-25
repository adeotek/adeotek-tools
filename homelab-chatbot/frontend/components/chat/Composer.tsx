"use client";

import { useState } from "react";

export function Composer({
  onSend,
  disabled,
}: {
  onSend: (text: string) => void;
  disabled: boolean;
}) {
  const [value, setValue] = useState("");

  function submit() {
    const trimmed = value.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setValue("");
  }

  return (
    <div className="border-t border-neutral-800 bg-neutral-900 p-3">
      <textarea
        aria-label="message"
        value={value}
        disabled={disabled}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            submit();
          }
        }}
        rows={2}
        placeholder="Ask about the home lab…"
        className="w-full resize-none rounded border border-neutral-700 bg-neutral-950 p-2 text-sm outline-none focus:border-blue-500"
      />
      <div className="mt-2 flex justify-end">
        <button
          onClick={submit}
          disabled={disabled || !value.trim()}
          className="rounded bg-blue-600 px-4 py-1.5 text-sm font-medium disabled:opacity-50"
        >
          Send
        </button>
      </div>
    </div>
  );
}
