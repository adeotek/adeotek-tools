import { useState } from 'react'
import { CornerDownLeft } from 'lucide-react'
import { useResizableHeight } from '../hooks/useResizableHeight'

interface ChatInputProps {
  onSend: (text: string) => void
  disabled?: boolean
}

export default function ChatInput({ onSend, disabled }: ChatInputProps) {
  const [value, setValue] = useState('')
  const { height: inputHeight, onDragStart } = useResizableHeight(72, 52, 400)

  function submit() {
    const text = value.trim()
    if (!text) return
    onSend(text)
    setValue('')
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  return (
    <div className="border-t border-border-subtle bg-bg-surface flex-shrink-0">
      {/* Resize handle — drag up/down to change input area height */}
      <div
        onMouseDown={onDragStart}
        className="h-1.5 cursor-ns-resize select-none hover:bg-accent/10 active:bg-accent/20 transition-colors"
      />
      <div style={{ height: inputHeight }} className="px-4 py-2 flex gap-2 items-end overflow-hidden">
        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={onKeyDown}
          disabled={disabled}
          autoFocus
          placeholder="Ask Claude Code anything… (Shift+Enter for new line)"
          className="flex-1 self-stretch bg-bg-panel border border-border rounded-md px-3 py-2 text-xs text-text-primary placeholder-text-dim focus:outline-none focus:border-accent resize-none leading-relaxed disabled:opacity-40 overflow-y-auto"
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
    </div>
  )
}
