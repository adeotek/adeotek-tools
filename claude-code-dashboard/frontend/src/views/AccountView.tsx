import { AlertTriangle } from 'lucide-react'
import InfoCard from '../components/InfoCard'
import { useAccount } from '../hooks/useAccount'

export default function AccountView() {
  const { data, loading, error } = useAccount()

  if (loading) {
    return <div className="p-6 text-text-muted text-sm animate-pulse">Loading account info…</div>
  }

  if (error) {
    return <div className="p-6 text-status-red text-sm">Failed to load account info: {error}</div>
  }

  if (!data?.claudeInstalled) {
    return (
      <div className="p-6 flex items-start gap-3">
        <AlertTriangle size={16} className="text-accent mt-0.5 flex-shrink-0" />
        <div>
          <p className="text-text-primary text-sm">Claude Code not found</p>
          <p className="text-text-muted text-xs mt-1">
            Install and authenticate Claude Code on this host, then reload the dashboard.
            <br />
            Check the <code className="text-accent">CLAUDE_BIN</code> env var if the binary is in a non-standard location.
          </p>
        </div>
      </div>
    )
  }

  const authColor = data.authStatus === 'authenticated' ? 'text-status-green' : 'text-status-red'

  return (
    <div className="p-6 space-y-6 max-w-2xl">
      <h2 className="text-text-secondary text-xs uppercase tracking-widest">Account</h2>
      <div className="grid grid-cols-2 gap-3">
        <InfoCard label="Version" value={data.version} />
        <InfoCard label="Model" value={data.model} accent />
        <InfoCard label="OS" value={data.os} sub={`Node ${data.nodeVersion}`} />
        <div className="bg-bg-surface border border-border-subtle rounded-md p-4">
          <div className="text-text-muted text-xs uppercase tracking-widest mb-2">Auth</div>
          <div className={`text-sm font-medium ${authColor}`}>
            {data.authStatus}
          </div>
        </div>
      </div>
    </div>
  )
}
