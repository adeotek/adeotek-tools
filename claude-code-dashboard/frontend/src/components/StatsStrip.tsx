import { Loader2, Settings } from 'lucide-react'
import { NavLink } from 'react-router-dom'
import type { AccountInfo } from '../hooks/useAccount'
import type { UsageData } from '../hooks/useUsage'

interface StatsStripProps {
  account: AccountInfo | null
  usage: UsageData | null
  activeSessions: number
  loading?: boolean
}

function formatTimeUntil(isoString: string | null): string | null {
  if (!isoString) return null
  const diffMs = new Date(isoString).getTime() - Date.now()
  if (diffMs <= 0) return null
  const totalMin = Math.ceil(diffMs / 60_000)
  if (totalMin < 60) return `${totalMin}m`
  const h = Math.floor(totalMin / 60)
  const m = totalMin % 60
  if (h < 24) return m > 0 ? `${h}h ${m}m` : `${h}h`
  const d = Math.floor(h / 24)
  const rh = h % 24
  return rh > 0 ? `${d}d ${rh}h` : `${d}d`
}

function pctColor(pct: number): string {
  if (pct >= 80) return 'text-status-red'
  if (pct >= 50) return 'text-yellow-400'
  return 'text-status-green'
}

function formatSubscription(s: string): string {
  if (s.startsWith('max')) return 'Max'
  if (s === 'pro') return 'Pro'
  if (s === 'enterprise') return 'Enterprise'
  return s.charAt(0).toUpperCase() + s.slice(1)
}

interface ChipProps {
  label: string
  value: string
  valueClass?: string
  reset?: string | null
}

function Chip({ label, value, valueClass = 'text-text-primary', reset }: ChipProps) {
  return (
    <span className="flex items-center gap-1.5 px-2.5 py-1 bg-bg-elevated rounded border border-border text-xs">
      <span className="text-text-muted uppercase tracking-wider">{label}</span>
      <span className={`font-medium ${valueClass}`}>{value}</span>
      {reset && (
        <>
          <span className="text-text-dim">·</span>
          <span className="text-text-dim">{reset}</span>
        </>
      )}
    </span>
  )
}

function SkeletonChip() {
  return <span className="w-16 h-6 bg-bg-elevated border border-border rounded animate-pulse" />
}

export default function StatsStrip({ account, usage, activeSessions, loading }: StatsStripProps) {
  const rl = usage?.rateLimits ?? null
  const isInitialLoad = loading && account === null

  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b border-border-subtle bg-bg-surface flex-shrink-0">
      <span className="text-accent font-semibold text-xs tracking-wide flex-shrink-0 mr-1">Claude Code</span>
      <div className="flex flex-wrap gap-2 items-center flex-1 min-w-0">
        {isInitialLoad && (
          <>
            <SkeletonChip />
            <SkeletonChip />
            <SkeletonChip />
            <SkeletonChip />
          </>
        )}
        {account?.version && (
          <Chip
            label="ver"
            value={account.version.startsWith('v') ? account.version : `v${account.version}`}
          />
        )}

        {rl?.subscriptionType && (
          <Chip label="plan" value={formatSubscription(rl.subscriptionType)} valueClass="text-accent font-medium" />
        )}

        {rl != null && (
          <Chip
            label="5h"
            value={`${rl.fiveHourPct}%`}
            valueClass={pctColor(rl.fiveHourPct)}
            reset={formatTimeUntil(rl.fiveHourResetsAt)}
          />
        )}

        {rl != null && (
          <Chip
            label="wk"
            value={`${rl.sevenDayPct}%`}
            valueClass={pctColor(rl.sevenDayPct)}
            reset={formatTimeUntil(rl.sevenDayResetsAt)}
          />
        )}

        {rl?.monthlyPct != null && (
          <Chip label="mo" value={`${rl.monthlyPct}%`} valueClass={pctColor(rl.monthlyPct)} />
        )}

        {/* ext: non-enterprise extra credits (mutually exclusive with ent) */}
        {rl?.extraEnabled && !rl.isEnterprise && (
          <Chip
            label="ext"
            value={`$${(rl.extraUsedCents / 100).toFixed(2)}/${rl.extraLimitCents == null ? '∞' : `$${(rl.extraLimitCents / 100).toFixed(2)}`}`}
            valueClass={
              rl.extraLimitCents != null && rl.extraLimitCents > 0
                ? pctColor(Math.round((rl.extraUsedCents / rl.extraLimitCents) * 100))
                : 'text-text-primary'
            }
          />
        )}

        {/* ent: enterprise usage (mutually exclusive with ext) */}
        {rl?.extraEnabled && rl.isEnterprise && (
          <Chip
            label="ent"
            value={`$${(rl.extraUsedCents / 100).toFixed(2)}/${rl.extraLimitCents == null ? '∞' : `$${(rl.extraLimitCents / 100).toFixed(2)}`}`}
            valueClass={
              rl.extraLimitCents != null && rl.extraLimitCents > 0
                ? pctColor(Math.round((rl.extraUsedCents / rl.extraLimitCents) * 100))
                : 'text-text-primary'
            }
          />
        )}

        {!isInitialLoad && (
          <Chip label="sessions" value={String(activeSessions)} />
        )}

        {/* Background refresh spinner — only when data is already shown */}
        {loading && !isInitialLoad && (
          <Loader2 size={12} className="text-text-dim animate-spin ml-1" />
        )}
      </div>

      <NavLink
        to="/settings"
        title="Settings"
        className={({ isActive }) =>
          'flex-shrink-0 flex items-center justify-center w-7 h-7 rounded transition-colors ' +
          (isActive ? 'text-accent' : 'text-text-dim hover:text-text-muted hover:bg-bg-elevated')
        }
      >
        <Settings size={14} strokeWidth={1.5} />
      </NavLink>
    </div>
  )
}
