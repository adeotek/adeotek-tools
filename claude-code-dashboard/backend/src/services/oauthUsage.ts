import fs from 'fs'
import path from 'path'
import os from 'os'
import { db } from '../db/schema'

export interface OAuthUsageData {
  fiveHourPct: number
  fiveHourResetsAt: string | null
  sevenDayPct: number
  sevenDayResetsAt: string | null
  monthlyPct: number | null
  monthlyResetsAt: string | null
  extraEnabled: boolean
  extraUsedCents: number
  extraLimitCents: number | null   // null = unlimited Enterprise
  currency: string
  subscriptionType: string | null
  isEnterprise: boolean
}

const CACHE_TTL_MS = 120_000          // 2 min
const ERROR_TTL_MS = 60 * 60_000     // 1 h back-off on auth/permission errors

let cache: { data: OAuthUsageData | null; expiresAt: number } | null = null
let lastGoodData: OAuthUsageData | null = null

const OAUTH_CACHE_KEY = 'usage'

function loadFromDisk(): { data: OAuthUsageData; cachedAt: number } | null {
  try {
    const row = db.prepare('SELECT data, cached_at FROM oauth_cache WHERE key = ?').get(OAUTH_CACHE_KEY) as
      | { data: string; cached_at: number }
      | undefined
    if (!row) return null
    return { data: JSON.parse(row.data) as OAuthUsageData, cachedAt: row.cached_at }
  } catch {
    return null
  }
}

function saveToDisk(data: OAuthUsageData, cachedAt: number): void {
  try {
    db.prepare('INSERT OR REPLACE INTO oauth_cache (key, data, cached_at) VALUES (?, ?, ?)').run(
      OAUTH_CACHE_KEY, JSON.stringify(data), cachedAt,
    )
  } catch {
    // non-fatal — disk write failure doesn't break the service
  }
}

// Warm in-memory cache from disk on startup to survive restarts
const _disk = loadFromDisk()
if (_disk) {
  lastGoodData = _disk.data
  if (Date.now() < _disk.cachedAt + CACHE_TTL_MS) {
    cache = { data: _disk.data, expiresAt: _disk.cachedAt + CACHE_TTL_MS }
  }
}

function readCredentials(): { token: string; subscriptionType: string | null } | null {
  try {
    const raw = fs.readFileSync(
      path.join(os.homedir(), '.claude', '.credentials.json'), 'utf8',
    )
    const creds = JSON.parse(raw) as Record<string, unknown>
    const oauth = creds.claudeAiOauth as Record<string, unknown> | undefined
    const token = typeof oauth?.accessToken === 'string' ? oauth.accessToken : null
    if (!token) return null
    const subscriptionType = typeof oauth?.subscriptionType === 'string' ? oauth.subscriptionType : null
    return { token, subscriptionType }
  } catch {
    return null
  }
}

export async function fetchOAuthUsage(): Promise<OAuthUsageData | null> {
  if (cache && Date.now() < cache.expiresAt) return cache.data

  const creds = readCredentials()
  if (!creds) {
    cache = { data: null, expiresAt: Date.now() + ERROR_TTL_MS }
    return null
  }

  try {
    const res = await fetch('https://api.anthropic.com/api/oauth/usage', { // nosec — CA bundle set via NODE_EXTRA_CA_CERTS on hosts that need it
      headers: {
        Authorization: `Bearer ${creds.token}`,
        'Content-Type': 'application/json',
        'anthropic-beta': 'oauth-2025-04-20',
      },
      signal: AbortSignal.timeout(5_000),
    })

    if (!res.ok) {
      if (res.status === 401 || res.status === 403) {
        // Credentials invalid — evict cache entirely
        cache = { data: null, expiresAt: Date.now() + ERROR_TTL_MS }
        lastGoodData = null
        return null
      }
      // Transient error (429, 5xx, etc.) — serve stale data if available, retry after TTL
      cache = { data: lastGoodData, expiresAt: Date.now() + CACHE_TTL_MS }
      return lastGoodData
    }

    const json = await res.json() as Record<string, unknown>
    const fiveHour = json.five_hour as Record<string, unknown> | undefined
    const sevenDay = json.seven_day as Record<string, unknown> | undefined
    const monthly = json.monthly as Record<string, unknown> | undefined
    const extra = json.extra_usage as Record<string, unknown> | undefined

    const data: OAuthUsageData = {
      fiveHourPct: Math.round(Number(fiveHour?.utilization ?? 0)),
      fiveHourResetsAt: typeof fiveHour?.resets_at === 'string' ? fiveHour.resets_at : null,
      sevenDayPct: Math.round(Number(sevenDay?.utilization ?? 0)),
      sevenDayResetsAt: typeof sevenDay?.resets_at === 'string' ? sevenDay.resets_at : null,
      monthlyPct: monthly?.utilization != null ? Math.round(Number(monthly.utilization)) : null,
      monthlyResetsAt: typeof monthly?.resets_at === 'string' ? monthly.resets_at : null,
      extraEnabled: extra?.is_enabled === true,
      extraUsedCents: Math.round(Number(extra?.used_credits ?? 0)),
      extraLimitCents: extra?.monthly_limit == null ? null : Math.round(Number(extra.monthly_limit)),
      currency: typeof extra?.currency === 'string' ? extra.currency : 'USD',
      subscriptionType: creds.subscriptionType,
      isEnterprise: creds.subscriptionType === 'enterprise',
    }

    const now = Date.now()
    lastGoodData = data
    saveToDisk(data, now)
    cache = { data, expiresAt: now + CACHE_TTL_MS }
    return data
  } catch (err) {
    console.warn('[oauthUsage] fetch failed:', (err as Error).message)
    // Network error — serve stale data if available, retry after TTL
    cache = { data: lastGoodData, expiresAt: Date.now() + CACHE_TTL_MS }
    return lastGoodData
  }
}
