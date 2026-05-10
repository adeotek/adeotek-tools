import { useState, useEffect } from 'react'
import { Plus, List, Square } from 'lucide-react'
import { useSession } from '../context/SessionContext'

interface SessionHeaderProps {
  onNewSession: () => void
  onStopSession: () => void
  onSessionsList: () => void
  totalTokens: number
  sessionStartedAt: number | null
}

function formatModelName(model: string): string {
  // Strip "claude-" prefix and any date suffix like "-20251001"
  const s = model.replace(/^claude-/, '').replace(/-\d{8}$/, '')
  // New naming: "sonnet-4-6", "opus-4-7", "haiku-4-5"
  const m = s.match(/^(opus|sonnet|haiku)-(\d+)-(\d+)/)
  if (m) return `${m[1].charAt(0).toUpperCase() + m[1].slice(1)} ${m[2]}.${m[3]}`
  // Old naming: "3-5-sonnet", "3-opus"
  const m2 = s.match(/^(\d+)-(?:(\d+)-)?(\w+)$/)
  if (m2) {
    const family = m2[3].charAt(0).toUpperCase() + m2[3].slice(1)
    return m2[2] ? `${family} ${m2[1]}.${m2[2]}` : `${family} ${m2[1]}`
  }
  // Bare short name: "sonnet", "opus", "haiku"
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
}: SessionHeaderProps) {
  const { state } = useSession()
  const [, setTick] = useState(0)

  // Tick only while Claude is actively running — idle time doesn't count
  useEffect(() => {
    if (state.wsState !== 'running') return
    const timer = setInterval(() => setTick((t) => t + 1), 1000)
    return () => clearInterval(timer)
  }, [state.wsState])

  const workingMs =
    state.workingTimeMs +
    (state.runningStartedAt != null ? Date.now() - state.runningStartedAt : 0)

  return (
    <div className="px-3 py-2 border-b border-border-subtle bg-bg-surface flex items-center gap-3 flex-shrink-0 flex-wrap">
      {/* Status + workdir */}
      <div className="flex items-center gap-2 min-w-0">
        <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
          state.wsState === 'running' ? 'bg-status-green' :
          state.wsState === 'idle'    ? 'bg-status-green/40' :
          state.wsState === 'error'   ? 'bg-status-red' : 'bg-text-dim'
        }`} />
        {state.workdir && (
          <span className="text-text-secondary text-xs bg-bg-elevated border border-border-subtle px-2 py-0.5 rounded truncate max-w-xs">
            {state.workdir}
          </span>
        )}
        {state.model && (
          <span className="text-accent text-xs font-medium hidden sm:block">{formatModelName(state.model)}</span>
        )}
      </div>

      {/* Session stat chips */}
      <div className="flex items-center gap-3 text-xs border-l border-border-subtle pl-3">
        {sessionStartedAt != null && (
          <StatChip label="created" value={formatCreatedAt(sessionStartedAt)} />
        )}
        {workingMs > 0 && (
          <StatChip label="dur" value={formatDuration(workingMs)} valueClass="text-status-green" />
        )}
        <StatChip label="tokens" value={formatTokens(totalTokens)} />
      </div>

      {/* Action buttons */}
      <div className="flex items-center gap-2 ml-auto">
        <button
          onClick={onNewSession}
          className="flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
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
