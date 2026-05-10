import { Plus, Play, Square, Trash2 } from 'lucide-react'
import type { Session } from '../hooks/useDashboard'

interface SessionListProps {
  sessions: Session[]
  onResume: (session: Session) => void
  onStop: (sessionId: string) => void
  onDelete: (sessionId: string) => void
  onNewSession: () => void
}

function formatRelativeTime(ts: number): string {
  const diffMs = Date.now() - ts
  const diffSec = Math.floor(diffMs / 1000)
  if (diffSec < 60) return 'just now'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin}m ago`
  const diffH = Math.floor(diffMin / 60)
  if (diffH < 24) return `${diffH}h ago`
  const diffD = Math.floor(diffH / 24)
  if (diffD < 30) return `${diffD} day${diffD === 1 ? '' : 's'} ago`
  const diffW = Math.floor(diffD / 7)
  if (diffD < 60) return `${diffW} week${diffW === 1 ? '' : 's'} ago`
  const diffMo = Math.floor(diffD / 30)
  return `${diffMo} month${diffMo === 1 ? '' : 's'} ago`
}

function lastSegment(path: string): string {
  const trimmed = path.replace(/\/$/, '')
  const idx = trimmed.lastIndexOf('/')
  return idx >= 0 ? trimmed.slice(idx + 1) : trimmed
}

async function deleteSession(sessionId: string): Promise<void> {
  await fetch(`/api/sessions/${sessionId}`, { method: 'DELETE' })
}

export default function SessionList({
  sessions,
  onResume,
  onStop,
  onDelete,
  onNewSession,
}: SessionListProps) {
  async function handleDelete(sessionId: string) {
    if (!window.confirm('Delete this session? This cannot be undone.')) return
    try {
      await deleteSession(sessionId)
      onDelete(sessionId)
    } catch {
      // ignore errors — parent can refresh
    }
  }

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-2.5 border-b border-border-subtle bg-bg-surface flex items-center justify-between flex-shrink-0">
        <span className="text-text-muted text-xs uppercase tracking-widest">Sessions</span>
        <button
          onClick={onNewSession}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-accent text-black hover:bg-accent-hover rounded transition-colors"
        >
          <Plus size={11} />
          New Session
        </button>
      </div>

      {/* Session rows */}
      <div className="flex-1 overflow-y-auto">
        {sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-4 p-6">
            <p className="text-text-muted text-sm">No sessions yet</p>
            <button
              onClick={onNewSession}
              className="flex items-center gap-2 px-4 py-2 text-sm bg-accent text-black hover:bg-accent-hover rounded transition-colors"
            >
              <Plus size={14} />
              Start New Session
            </button>
          </div>
        ) : (
          sessions.map((session) => (
            <div
              key={session.id}
              className="flex items-center gap-4 px-4 py-3 border-b border-border-subtle hover:bg-bg-elevated transition-colors"
            >
              {/* Left: workdir info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-text-primary text-sm font-medium truncate">
                    {lastSegment(session.workdir)}
                  </span>
                  {session.is_active ? (
                    <span className="flex-shrink-0 px-1.5 py-0.5 text-xs bg-status-green/10 text-status-green border border-status-green/20 rounded">
                      active
                    </span>
                  ) : (
                    <span className="flex-shrink-0 px-1.5 py-0.5 text-xs bg-bg-panel text-text-dim border border-border-subtle rounded">
                      ended
                    </span>
                  )}
                </div>
                <div className="text-text-muted text-xs mt-0.5 truncate">{session.workdir}</div>
              </div>

              {/* Middle: time + message count */}
              <div className="text-right flex-shrink-0 hidden sm:block">
                <div className="text-text-secondary text-xs">{formatRelativeTime(session.started_at)}</div>
                <div className="text-text-muted text-xs mt-0.5">{session.message_count} messages</div>
              </div>

              {/* Right: actions */}
              <div className="flex items-center gap-1.5 flex-shrink-0">
                <button
                  onClick={() => onResume(session)}
                  title={session.is_active ? 'Reconnect' : 'Resume'}
                  className="flex items-center justify-center px-2.5 py-1.5 text-xs border border-accent text-accent hover:bg-accent hover:text-black rounded transition-colors"
                >
                  <Play size={11} />
                </button>
                <button
                  onClick={() => onStop(session.id)}
                  disabled={!session.is_active}
                  title="Stop session"
                  className="flex items-center justify-center px-2.5 py-1.5 text-xs border rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed border-status-red text-status-red enabled:hover:bg-status-red enabled:hover:text-black"
                >
                  <Square size={11} />
                </button>
                <button
                  onClick={() => handleDelete(session.id)}
                  title="Delete session"
                  className="flex items-center gap-1 px-2.5 py-1.5 text-xs border border-status-red text-status-red hover:bg-status-red hover:text-black rounded transition-colors"
                >
                  <Trash2 size={11} />
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
