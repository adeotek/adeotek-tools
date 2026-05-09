import { useState, useEffect } from 'react'

export interface AccountInfo {
  claudeInstalled: boolean
  version: string | null
  model: string | null
  os: string
  nodeVersion: string
  authStatus: 'authenticated' | 'unauthenticated' | 'unknown'
}

export function useAccount() {
  const [data, setData] = useState<AccountInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/account')
      .then((r) => r.json())
      .then((d: AccountInfo) => { setData(d); setLoading(false) })
      .catch((e: Error) => { setError(e.message); setLoading(false) })
  }, [])

  return { data, loading, error }
}
