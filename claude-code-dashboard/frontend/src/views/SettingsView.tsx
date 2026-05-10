import { useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { NavLink } from 'react-router-dom'

const KEY = 'dashboard_anthropic_api_key'

export default function SettingsView() {
  const [apiKey, setApiKey] = useState(() => localStorage.getItem(KEY) ?? '')
  const [saved, setSaved] = useState(false)

  function handleSave() {
    if (apiKey) localStorage.setItem(KEY, apiKey)
    else localStorage.removeItem(KEY)
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
