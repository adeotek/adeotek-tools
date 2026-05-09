import { Plus } from 'lucide-react'
import { useSession } from '../context/SessionContext'

interface SessionHeaderProps {
  onNewSession: () => void
}

export default function SessionHeader({ onNewSession }: SessionHeaderProps) {
  const { state } = useSession()

  return (
    <div className="px-4 py-2 border-b border-border-subtle bg-bg-surface flex items-center gap-3 flex-shrink-0">
      <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
        state.wsState === 'running' ? 'bg-status-green' :
        state.wsState === 'idle' ? 'bg-text-muted' :
        state.wsState === 'error' ? 'bg-status-red' : 'bg-text-dim'
      }`} />
      <span className="text-text-muted text-xs uppercase tracking-widest flex-shrink-0">Session</span>
      {state.workdir && (
        <span className="text-text-secondary text-xs bg-bg-elevated border border-border-subtle px-2 py-0.5 rounded truncate max-w-xs">
          {state.workdir}
        </span>
      )}
      {state.model && (
        <span className="text-text-dim text-xs ml-1 hidden sm:block">{state.model}</span>
      )}
      <button
        onClick={onNewSession}
        className="ml-auto flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
      >
        <Plus size={11} />
        New session
      </button>
    </div>
  )
}
