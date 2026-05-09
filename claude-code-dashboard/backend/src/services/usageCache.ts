import { db } from '../db/schema'
import { parseLocalUsage, type DayUsage } from './localLogs'
import { fetchApiUsage } from './anthropicApi'

const CACHE_TTL_MS = 60 * 60 * 1000 // 1 hour

function isFresh(cachedAt: number): boolean {
  return Date.now() - cachedAt < CACHE_TTL_MS
}

export interface UsageResult {
  days: DayUsage[]
  totals: { inputTokens: number; outputTokens: number; sessions: number; costUsd: number }
  sources: string[]
}

export async function getUsage(month: string): Promise<UsageResult> {
  const cached = db
    .prepare('SELECT * FROM usage_cache WHERE date LIKE ? AND cached_at > ?')
    .all(`${month}%`, Date.now() - CACHE_TTL_MS) as Array<{
    date: string; input_tokens: number; output_tokens: number; cost_usd: number; source: string; cached_at: number
  }>

  if (cached.length > 0 && cached.every((r) => isFresh(r.cached_at))) {
    return buildResult(
      cached.map((r) => ({
        date: r.date,
        inputTokens: r.input_tokens,
        outputTokens: r.output_tokens,
        costUsd: r.cost_usd,
      })),
      month,
      [...new Set(cached.map((r) => r.source))],
    )
  }

  const [localDays, apiDays] = await Promise.all([
    Promise.resolve(parseLocalUsage(month)),
    fetchApiUsage(month),
  ])

  const merged = mergeDays(localDays, apiDays)
  const sources = ['local', ...(apiDays.length > 0 ? ['api'] : [])]

  const upsert = db.prepare(`
    INSERT INTO usage_cache (date, input_tokens, output_tokens, cost_usd, source, cached_at)
    VALUES (@date, @input_tokens, @output_tokens, @cost_usd, @source, @cached_at)
    ON CONFLICT(date) DO UPDATE SET
      input_tokens = excluded.input_tokens,
      output_tokens = excluded.output_tokens,
      cost_usd = excluded.cost_usd,
      source = excluded.source,
      cached_at = excluded.cached_at
  `)

  const insertMany = db.transaction((rows: DayUsage[]) => {
    for (const row of rows) {
      upsert.run({
        date: row.date,
        input_tokens: row.inputTokens,
        output_tokens: row.outputTokens,
        cost_usd: row.costUsd,
        source: sources.join('+'),
        cached_at: Date.now(),
      })
    }
  })
  insertMany(merged)

  return buildResult(merged, month, sources)
}

function mergeDays(local: DayUsage[], api: DayUsage[]): DayUsage[] {
  const map = new Map<string, DayUsage>()
  for (const d of local) map.set(d.date, { ...d })
  for (const d of api) {
    const existing = map.get(d.date)
    if (existing) {
      // API is authoritative for cost; local provides session-level token detail
      existing.costUsd = d.costUsd
    } else {
      map.set(d.date, { ...d })
    }
  }
  return [...map.values()].sort((a, b) => a.date.localeCompare(b.date))
}

function buildResult(days: DayUsage[], month: string, sources: string[]): UsageResult {
  const sessions = (db
    .prepare(`SELECT COUNT(*) as count FROM sessions WHERE started_at >= ? AND started_at < ?`)
    .get(
      new Date(`${month}-01`).getTime(),
      new Date(`${month}-01`).getTime() + 32 * 24 * 60 * 60 * 1000,
    ) as { count: number }).count

  const totals = days.reduce(
    (acc, d) => ({
      inputTokens: acc.inputTokens + d.inputTokens,
      outputTokens: acc.outputTokens + d.outputTokens,
      costUsd: acc.costUsd + d.costUsd,
      sessions: acc.sessions,
    }),
    { inputTokens: 0, outputTokens: 0, costUsd: 0, sessions },
  )

  return { days, totals, sources }
}
