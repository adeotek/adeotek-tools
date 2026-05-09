import Anthropic from '@anthropic-ai/sdk'
import type { DayUsage } from './localLogs'

let client: Anthropic | null = null

function getClient(): Anthropic | null {
  if (!process.env.ANTHROPIC_API_KEY) return null
  if (!client) client = new Anthropic({ apiKey: process.env.ANTHROPIC_API_KEY })
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
