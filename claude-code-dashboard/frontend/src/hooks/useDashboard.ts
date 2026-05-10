import { useState, useEffect, useCallback } from 'react'
import type { AccountInfo } from './useAccount'
import type { UsageData } from './useUsage'

export interface Session {
  id: string
  workdir: string
  model: string | null
  started_at: number
  ended_at: number | null
  is_active: boolean
  message_count: number
}

export interface DashboardData {
  account: AccountInfo | null
  usage: UsageData | null
  sessions: Session[]
  activeSessions: number
  loading: boolean
  error: string | null
}

export function useDashboard(month?: string): DashboardData & { refresh: () => void } {
  const [account, setAccount] = useState<AccountInfo | null>(null)
  const [usage, setUsage] = useState<UsageData | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const target = month ?? new Date().toISOString().slice(0, 7)

  const fetchAll = useCallback(() => {
    setLoading(true)
    setError(null)

    Promise.all([
      fetch('/api/account').then((r) => r.json() as Promise<AccountInfo>),
      fetch(`/api/usage?month=${target}`).then((r) => r.json() as Promise<UsageData>),
      fetch('/api/sessions').then((r) => r.json() as Promise<Session[]>),
    ])
      .then(([acc, usg, sess]) => {
        setAccount(acc)
        setUsage(usg)
        setSessions(Array.isArray(sess) ? sess : [])
        setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }, [target])

  useEffect(() => {
    fetchAll()
    const timer = setInterval(fetchAll, 60_000)
    return () => clearInterval(timer)
  }, [fetchAll])

  const activeSessions = sessions.filter((s) => s.is_active).length

  return { account, usage, sessions, activeSessions, loading, error, refresh: fetchAll }
}
