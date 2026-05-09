import fs from 'fs'
import path from 'path'
import os from 'os'

const CLAUDE_HOME = path.join(os.homedir(), '.claude')
const PROJECTS_DIR = path.join(CLAUDE_HOME, 'projects')

interface LogEntry {
  type: string
  costUSD?: number
  usage?: {
    input_tokens?: number
    output_tokens?: number
    cache_creation_input_tokens?: number
    cache_read_input_tokens?: number
  }
  timestamp?: string
}

export interface DayUsage {
  date: string
  inputTokens: number
  outputTokens: number
  costUsd: number
}

function parseDate(ts: string | undefined): string | null {
  if (!ts) return null
  const d = new Date(ts)
  if (isNaN(d.getTime())) return null
  return d.toISOString().slice(0, 10)
}

function walkJsonl(dir: string): string[] {
  const files: string[] = []
  if (!fs.existsSync(dir)) return files
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) files.push(...walkJsonl(full))
    else if (entry.isFile() && entry.name.endsWith('.jsonl')) files.push(full)
  }
  return files
}

export function parseLocalUsage(month: string): DayUsage[] {
  const byDate = new Map<string, DayUsage>()

  for (const file of walkJsonl(PROJECTS_DIR)) {
    const lines = fs.readFileSync(file, 'utf8').split('\n')
    for (const line of lines) {
      if (!line.trim()) continue
      let entry: LogEntry
      try { entry = JSON.parse(line) as LogEntry } catch { continue }

      if (entry.type !== 'assistant' || !entry.usage) continue
      const date = parseDate(entry.timestamp)
      if (!date || !date.startsWith(month)) continue

      const cur = byDate.get(date) ?? { date, inputTokens: 0, outputTokens: 0, costUsd: 0 }
      cur.inputTokens += (entry.usage.input_tokens ?? 0) + (entry.usage.cache_read_input_tokens ?? 0)
      cur.outputTokens += entry.usage.output_tokens ?? 0
      cur.costUsd += entry.costUSD ?? 0
      byDate.set(date, cur)
    }
  }

  return [...byDate.values()].sort((a, b) => a.date.localeCompare(b.date))
}
