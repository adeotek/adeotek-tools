import { useState, useEffect } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { NavLink } from 'react-router-dom'

const API_KEY = 'dashboard_anthropic_api_key'

export default function SettingsView() {
  const [apiKey, setApiKey] = useState(() => localStorage.getItem(API_KEY) ?? '')
  const [bypassPermissions, setBypassPermissions] = useState(true)
  const [saved, setSaved] = useState(false)
  const [settingsLoaded, setSettingsLoaded] = useState(false)

  useEffect(() => {
    fetch('/api/settings')
      .then((r) => r.json() as Promise<Record<string, string>>)
      .then((data) => {
        setBypassPermissions(data.bypass_permissions !== 'false')
        setSettingsLoaded(true)
      })
      .catch(() => setSettingsLoaded(true))
  }, [])

  async function handleSave() {
    if (apiKey) localStorage.setItem(API_KEY, apiKey)
    else localStorage.removeItem(API_KEY)

    await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ bypass_permissions: String(bypassPermissions) }),
    }).catch(() => {})

    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="p-6 space-y-6 max-w-md">
      <NavLink
        to="/"
        className="inline-flex items-center gap-1.5 text-xs text-accent hover:text-accent-hover transition-colors"
      >
        <ArrowLeft size={13} strokeWidth={1.5} />
        Back to Dashboard
      </NavLink>

      <h2 className="text-text-secondary text-xs uppercase tracking-widest">Settings</h2>

      {/* API key */}
      <div className="space-y-2">
        <label className="text-text-secondary text-xs uppercase tracking-widest block">
          Anthropic API Key
        </label>
        <p className="text-text-muted text-xs">
          Used to fetch billing usage totals in the Usage view. Stored in localStorage only.
        </p>
        <input
          type="password"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="sk-ant-…"
          className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
        />
      </div>

      {/* Bypass permissions toggle */}
      <div className="space-y-2">
        <label className="text-text-secondary text-xs uppercase tracking-widest block">
          Tool Permissions
        </label>
        <p className="text-text-muted text-xs">
          When enabled, Claude Code runs with <code className="text-text-secondary">--dangerously-skip-permissions</code> — the same trust level as running <code className="text-text-secondary">claude</code> directly in your terminal. When disabled, writes outside the workspace are sandboxed.
        </p>
        <button
          onClick={() => setBypassPermissions((v) => !v)}
          disabled={!settingsLoaded}
          className={`flex items-center gap-2 px-3 py-1.5 text-xs rounded border transition-colors disabled:opacity-40 ${
            bypassPermissions
              ? 'border-status-green text-status-green bg-status-green/10'
              : 'border-border-subtle text-text-muted bg-bg-elevated'
          }`}
        >
          <span className={`w-2 h-2 rounded-full ${bypassPermissions ? 'bg-status-green' : 'bg-text-dim'}`} />
          {bypassPermissions ? 'Bypass permissions: on' : 'Bypass permissions: off'}
        </button>
      </div>

      <button
        onClick={handleSave}
        className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-black text-sm font-medium px-4 py-2 rounded-md transition-colors"
      >
        <Save size={14} />
        {saved ? 'Saved!' : 'Save'}
      </button>
    </div>
  )
}
