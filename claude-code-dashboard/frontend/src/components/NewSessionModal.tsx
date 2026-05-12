import { useState, useEffect } from 'react'
import { FolderOpen } from 'lucide-react'

interface NewSessionModalProps {
  onStart: (sessionId: string, workdir: string, name: string | null) => void
}

export default function NewSessionModal({ onStart }: NewSessionModalProps) {
  const [name, setName] = useState('')
  const [workdir, setWorkdir] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [dirs, setDirs] = useState<string[]>([])
  const [showDropdown, setShowDropdown] = useState(false)
  const [highlightedIndex, setHighlightedIndex] = useState(-1)

  // Derive parent path (up to and including last '/') and the current prefix being typed
  const lastSlash = workdir.lastIndexOf('/')
  const parentPath = lastSlash >= 0 ? workdir.slice(0, lastSlash + 1) : null
  const prefix = lastSlash >= 0 ? workdir.slice(lastSlash + 1) : ''

  // Fetch subdirectories whenever the parent path changes
  useEffect(() => {
    if (parentPath === null) { setDirs([]); return }
    const controller = new AbortController()
    fetch(`/api/directories?path=${encodeURIComponent(parentPath)}`, { signal: controller.signal })
      .then((r) => r.json() as Promise<string[]>)
      .then(setDirs)
      .catch(() => {})
    return () => controller.abort()
  }, [parentPath])

  // Filter fetched dirs by the current prefix (substring after last '/')
  const filtered = dirs.filter((d) =>
    d.slice(parentPath?.length ?? 0).toLowerCase().startsWith(prefix.toLowerCase())
  )

  function selectDir(dir: string) {
    // Append '/' so the user can keep navigating into subdirectories
    setWorkdir(dir + '/')
    setHighlightedIndex(-1)
  }

  function handleWorkdirKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (!showDropdown || filtered.length === 0) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlightedIndex((i) => Math.min(i + 1, filtered.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlightedIndex((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter' && highlightedIndex >= 0) {
      e.preventDefault()
      selectDir(filtered[highlightedIndex])
    } else if (e.key === 'Escape') {
      setShowDropdown(false)
      setHighlightedIndex(-1)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!workdir.trim()) return
    setLoading(true)
    setError(null)
    try {
      const body: { workdir: string; name?: string } = { workdir: workdir.trim() }
      if (name.trim()) body.name = name.trim()
      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error(await res.text())
      const { sessionId } = (await res.json()) as { sessionId: string }
      onStart(sessionId, workdir.trim(), name.trim() || null)
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
          <div className="space-y-1">
            <label className="text-text-dim text-xs">
              Name <span className="text-text-muted">(optional)</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. API refactor"
              className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
            />
          </div>

          <div className="space-y-1">
            <label className="text-text-dim text-xs">Working directory</label>
            <div className="relative">
              <input
                type="text"
                value={workdir}
                onChange={(e) => { setWorkdir(e.target.value); setHighlightedIndex(-1) }}
                onFocus={() => setShowDropdown(true)}
                onBlur={() => { setShowDropdown(false); setHighlightedIndex(-1) }}
                onKeyDown={handleWorkdirKeyDown}
                placeholder="/home/user/projects/my-app"
                autoFocus
                className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
              />
              {showDropdown && filtered.length > 0 && (
                <ul className="absolute z-10 w-full mt-1 bg-bg-panel border border-border rounded-md shadow-lg max-h-52 overflow-y-auto">
                  {filtered.map((dir, i) => (
                    <li key={dir}>
                      <button
                        type="button"
                        onMouseDown={(e) => { e.preventDefault(); selectDir(dir) }}
                        className={`w-full text-left px-3 py-1.5 text-xs font-mono truncate transition-colors ${
                          i === highlightedIndex
                            ? 'bg-bg-elevated text-accent'
                            : 'text-text-secondary hover:bg-bg-elevated hover:text-text-primary'
                        }`}
                      >
                        {dir.slice(parentPath?.length ?? 0)}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

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
