import { useState, useEffect } from 'react'
import { Plus, List, Square, Pencil } from 'lucide-react'
import { useSession } from '../context/SessionContext'

interface SessionHeaderProps {
  onNewSession: () => void
  onStopSession: () => void
  onSessionsList: () => void
  totalTokens: number
  sessionStartedAt: number | null
  sessionName?: string | null
  onRename?: (name: string) => Promise<void>
}

function formatModelName(model: string): string {
  const s = model.replace(/^claude-/, '').replace(/-\d{8}$/, '')
  const m = s.match(/^(opus|sonnet|haiku)-(\d+)-(\d+)/)
  if (m) return `${m[1].charAt(0).toUpperCase() + m[1].slice(1)} ${m[2]}.${m[3]}`
  const m2 = s.match(/^(\d+)-(?:(\d+)-)?(\w+)$/)
  if (m2) {
    const family = m2[3].charAt(0).toUpperCase() + m2[3].slice(1)
    return m2[2] ? `${family} ${m2[1]}.${m2[2]}` : `${family} ${m2[1]}`
  }
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(n)
}

function formatCreatedAt(ts: number): string {
  const d = new Date(ts)
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (sameDay) return time
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ' ' + time
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

interface StatChipProps { label: string; value: string; valueClass?: string }
function StatChip({ label, value, valueClass = 'text-text-secondary' }: StatChipProps) {
  return (
    <span className="flex items-center gap-1 text-xs">
      <span className="text-text-dim uppercase tracking-wider">{label}</span>
      <span className={`font-medium ${valueClass}`}>{value}</span>
    </span>
  )
}

export default function SessionHeader({
  onNewSession,
  onStopSession,
  onSessionsList,
  totalTokens,
  sessionStartedAt,
  sessionName,
  onRename,
}: SessionHeaderProps) {
  const { state } = useSession()
  const [, setTick] = useState(0)
  const [renaming, setRenaming] = useState(false)
  const [nameInput, setNameInput] = useState('')
  const [renameError, setRenameError] = useState<string | null>(null)
  const [renameSaving, setRenameSaving] = useState(false)

  useEffect(() => {
    if (state.wsState !== 'running') return
    const timer = setInterval(() => setTick((t) => t + 1), 1000)
    return () => clearInterval(timer)
  }, [state.wsState])

  const workingMs =
    state.workingTimeMs +
    (state.runningStartedAt != null ? Date.now() - state.runningStartedAt : 0)

  function startRename() {
    setNameInput(sessionName ?? '')
    setRenameError(null)
    setRenaming(true)
  }

  function cancelRename() {
    setRenaming(false)
    setRenameError(null)
  }

  async function saveRename() {
    if (!onRename) return
    setRenameSaving(true)
    setRenameError(null)
    try {
      await onRename(nameInput.trim())
      setRenaming(false)
    } catch (err) {
      setRenameError(err instanceof Error ? err.message : 'Failed to rename')
    } finally {
      setRenameSaving(false)
    }
  }

  function onRenameKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') { e.preventDefault(); saveRename() }
    if (e.key === 'Escape') cancelRename()
  }

  return (
    <div className="px-3 py-2 border-b border-border-subtle bg-bg-surface flex items-center gap-3 flex-shrink-0 flex-wrap">
      {renaming ? (
        <div className="flex items-center gap-1.5 min-w-0 flex-wrap">
          <input
            value={nameInput}
            onChange={(e) => setNameInput(e.target.value)}
            onKeyDown={onRenameKeyDown}
            autoFocus
            placeholder="Session name (optional)"
            className="bg-bg-panel border border-accent rounded px-2 py-0.5 text-xs text-text-primary placeholder-text-dim focus:outline-none w-44"
          />
          <button
            onClick={saveRename}
            disabled={renameSaving}
            className="text-xs text-accent hover:text-accent-hover disabled:opacity-40 transition-colors"
          >
            {renameSaving ? '…' : 'Save'}
          </button>
          <button
            onClick={cancelRename}
            className="text-xs text-text-dim hover:text-text-secondary transition-colors"
          >
            ✕
          </button>
          {renameError && <span className="text-status-red text-xs">{renameError}</span>}
        </div>
      ) : (
        <div className="flex items-center gap-2 min-w-0">
          {state.wsState === 'running' ? (
            <div className="w-3 h-3 rounded-full border-2 border-status-green border-t-transparent animate-spin flex-shrink-0" />
          ) : (
            <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
              state.wsState === 'idle'  ? 'bg-status-green' :
              state.wsState === 'error' ? 'bg-status-red' : 'bg-text-dim'
            }`} />
          )}
          {sessionName && (
            <span className="text-text-primary text-xs font-medium truncate max-w-[160px]">
              {sessionName}
            </span>
          )}
          {onRename && !renaming && (
            <button
              onClick={startRename}
              className="text-text-dim hover:text-accent transition-colors flex-shrink-0"
              title="Rename session"
            >
              <Pencil size={11} />
            </button>
          )}
          {state.workdir && (
            <span className="text-text-secondary text-xs bg-bg-elevated border border-border-subtle px-2 py-0.5 rounded truncate max-w-xs">
              {state.workdir}
            </span>
          )}
          {state.model && (
            <span className="text-accent text-xs font-medium hidden sm:block">{formatModelName(state.model)}</span>
          )}
        </div>
      )}

      <div className="flex items-center gap-3 text-xs border-l border-border-subtle pl-3">
        {sessionStartedAt != null && (
          <StatChip label="created" value={formatCreatedAt(sessionStartedAt)} />
        )}
        {workingMs > 0 && (
          <StatChip label="dur" value={formatDuration(workingMs)} valueClass="text-status-green" />
        )}
        <StatChip label="tokens" value={formatTokens(totalTokens)} />
      </div>

      <div className="flex items-center gap-2 ml-auto">
        <button
          onClick={onNewSession}
          className="flex items-center gap-1.5 text-text-muted hover:text-accent text-xs bg-bg-elevated border border-border-subtle hover:border-accent px-2 py-1 rounded transition-colors"
        >
          <Plus size={11} />
          New session
        </button>
        <button
          onClick={onStopSession}
          className="flex items-center gap-1.5 text-text-muted hover:text-status-red text-xs bg-bg-elevated border border-border-subtle hover:border-status-red px-2 py-1 rounded transition-colors"
        >
          <Square size={11} />
          Stop session
        </button>
        <button
          onClick={onSessionsList}
          className="flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
        >
          <List size={11} />
          Sessions list
        </button>
      </div>
    </div>
  )
}
