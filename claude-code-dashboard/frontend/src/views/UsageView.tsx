import { AlertTriangle } from 'lucide-react'
import UsageChart from '../components/UsageChart'
import StatCard from '../components/StatCard'
import { useUsage } from '../hooks/useUsage'

function formatTokens(n: number) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}k`
  return String(n)
}

export default function UsageView() {
  const { data, loading, error } = useUsage()
  const month = new Date().toLocaleString('default', { month: 'long', year: 'numeric' })

  if (loading) {
    return <div className="p-6 text-text-muted text-sm animate-pulse">Loading usage data…</div>
  }

  if (error) {
    return <div className="p-6 text-status-red text-sm">Failed to load usage: {error}</div>
  }

  const { days = [], totals, sources = [] } = data ?? {}
  const hasApiData = sources.includes('api')

  return (
    <div className="p-6 space-y-5 max-w-2xl">
      <div className="flex items-center justify-between">
        <h2 className="text-text-secondary text-xs uppercase tracking-widest">Usage · {month}</h2>
        {!hasApiData && (
          <div className="flex items-center gap-1.5 text-text-muted text-xs">
            <AlertTriangle size={11} className="text-accent" />
            billing data unavailable — set <code className="text-accent ml-0.5">ANTHROPIC_API_KEY</code>
          </div>
        )}
      </div>

      <div className="bg-bg-surface border border-border-subtle rounded-md p-4">
        <UsageChart days={days} />
      </div>

      {totals && (
        <div className="grid grid-cols-3 gap-3">
          <StatCard
            label="Tokens"
            value={formatTokens(totals.inputTokens + totals.outputTokens)}
            color="amber"
          />
          <StatCard label="Sessions" value={String(totals.sessions)} color="green" />
          <StatCard label="Est. cost" value={`$${totals.costUsd.toFixed(2)}`} color="blue" />
        </div>
      )}
    </div>
  )
}
