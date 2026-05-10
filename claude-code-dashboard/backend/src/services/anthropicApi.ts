import Anthropic from '@anthropic-ai/sdk'
import fs from 'fs'
import path from 'path'
import os from 'os'
import type { DayUsage } from './localLogs'

let client: Anthropic | null = null

function resolveApiKey(): string | null {
  if (process.env.ANTHROPIC_API_KEY) return process.env.ANTHROPIC_API_KEY

  // Fall back to ~/.claude/settings.json env block — where Claude Code users
  // typically store ANTHROPIC_API_KEY without it being in the shell environment.
  try {
    const settingsPath = path.join(os.homedir(), '.claude', 'settings.json')
    const settings = JSON.parse(fs.readFileSync(settingsPath, 'utf8')) as Record<string, unknown>
    const env = settings.env as Record<string, string> | undefined
    if (env?.ANTHROPIC_API_KEY) return env.ANTHROPIC_API_KEY
  } catch {
    // settings.json absent or malformed — no key available
  }

  return null
}

function getClient(): Anthropic | null {
  const apiKey = resolveApiKey()
  if (!apiKey) return null
  if (!client) client = new Anthropic({ apiKey })
  return client
}

export async function fetchApiUsage(month: string): Promise<DayUsage[]> {
  const api = getClient()
  if (!api) return []

  try {
    const [year, mon] = month.split('-').map(Number)
    const startDate = new Date(year, mon - 1, 1)
    const endDate = new Date(year, mon, 0)

    // @ts-ignore — usage endpoint may not be in current SDK types
    const response = await api.usage.monthly({
      start_time: startDate.toISOString(),
      end_time: endDate.toISOString(),
    })

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const items: any[] = response?.data ?? []
    return items.map((item) => ({
      date: String(item.date ?? '').slice(0, 10),
      inputTokens: Number(item.input_tokens ?? 0),
      outputTokens: Number(item.output_tokens ?? 0),
      costUsd: Number(item.cost_usd ?? 0),
    }))
  } catch {
    return []
  }
}
