import { useState, useRef } from 'react'
import { CornerDownLeft } from 'lucide-react'

interface ChatInputProps {
  onSend: (text: string) => void
  disabled?: boolean
}

export default function ChatInput({ onSend, disabled }: ChatInputProps) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  function submit() {
    const text = value.trim()
    if (!text) return
    onSend(text)
    setValue('')
    if (textareaRef.current) textareaRef.current.style.height = 'auto'
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  function onInput(e: React.FormEvent<HTMLTextAreaElement>) {
    const el = e.currentTarget
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }

  return (
    <div className="px-4 py-3 border-t border-border-subtle bg-bg-surface flex gap-2 items-end flex-shrink-0">
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={onKeyDown}
        onInput={onInput}
        disabled={disabled}
        autoFocus
        placeholder="Ask Claude Code anything… (Shift+Enter for new line)"
        rows={1}
        className="flex-1 bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent resize-none leading-relaxed disabled:opacity-40"
      />
      <button
        onClick={submit}
        disabled={disabled || !value.trim()}
        className="bg-accent hover:bg-accent-hover disabled:opacity-40 disabled:cursor-not-allowed text-black p-2 rounded-md transition-colors flex-shrink-0"
        title="Send (Enter)"
      >
        <CornerDownLeft size={14} />
      </button>
    </div>
  )
}
