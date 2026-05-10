import { useState } from 'react'
import { FolderOpen } from 'lucide-react'

interface NewSessionModalProps {
  onStart: (sessionId: string, workdir: string) => void
}

export default function NewSessionModal({ onStart }: NewSessionModalProps) {
  const [workdir, setWorkdir] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!workdir.trim()) return
    setLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workdir: workdir.trim() }),
      })
      if (!res.ok) throw new Error(await res.text())
      const { sessionId } = (await res.json()) as { sessionId: string }
      onStart(sessionId, workdir.trim())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session')
      setLoading(false)
    }
  }

  return (
    <div className="flex-1 flex items-center justify-center p-6">
      <div className="w-full max-w-md bg-bg-surface border border-border rounded-lg p-6 space-y-4">
        <div className="flex items-center gap-2">
          <FolderOpen size={16} className="text-accent" />
          <h3 className="text-text-primary text-sm font-medium">Start new session</h3>
        </div>
        <p className="text-text-muted text-xs">
          Enter the working directory where Claude Code will run.
        </p>
        <form onSubmit={handleSubmit} className="space-y-3">
          <input
            type="text"
            value={workdir}
            onChange={(e) => setWorkdir(e.target.value)}
            placeholder="/home/user/projects/my-app"
            autoFocus
            className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
          />
          {error && <p className="text-status-red text-xs">{error}</p>}
          <button
            type="submit"
            disabled={loading || !workdir.trim()}
            className="w-full bg-accent hover:bg-accent-hover disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium py-2 rounded-md transition-colors"
          >
            {loading ? 'Starting…' : 'Start session'}
          </button>
        </form>
      </div>
    </div>
  )
}
