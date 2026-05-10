import { useState, useEffect } from 'react'

export interface DayUsage {
  date: string
  inputTokens: number
  outputTokens: number
  costUsd: number
}

export interface OAuthUsageData {
  fiveHourPct: number
  fiveHourResetsAt: string | null
  sevenDayPct: number
  sevenDayResetsAt: string | null
  monthlyPct: number | null
  monthlyResetsAt: string | null
  extraEnabled: boolean
  extraUsedCents: number
  extraLimitCents: number | null
  currency: string
  subscriptionType: string | null
  isEnterprise: boolean
}

export interface UsageData {
  days: DayUsage[]
  totals: { inputTokens: number; outputTokens: number; sessions: number; costUsd: number }
  sources: string[]
  rateLimits: OAuthUsageData | null
}

export function useUsage(month?: string) {
  const [data, setData] = useState<UsageData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const target = month ?? new Date().toISOString().slice(0, 7)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/usage?month=${target}`)
      .then((r) => r.json())
      .then((d: UsageData) => { setData(d); setLoading(false) })
      .catch((e: Error) => { setError(e.message); setLoading(false) })
  }, [target])

  return { data, loading, error }
}
