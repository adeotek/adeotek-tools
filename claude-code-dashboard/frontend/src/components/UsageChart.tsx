import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { DayUsage } from '../hooks/useUsage'

interface UsageChartProps {
  days: DayUsage[]
}

function formatK(n: number) {
  return n >= 1000 ? `${(n / 1000).toFixed(0)}k` : String(n)
}

export default function UsageChart({ days }: UsageChartProps) {
  const data = days.map((d) => ({
    date: d.date.slice(5), // MM-DD
    tokens: d.inputTokens + d.outputTokens,
    cost: d.costUsd,
  }))

  return (
    <ResponsiveContainer width="100%" height={180}>
      <AreaChart data={data} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
        <defs>
          <linearGradient id="tokenGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#d97706" stopOpacity={0.25} />
            <stop offset="95%" stopColor="#d97706" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="#1e1e1e" />
        <XAxis
          dataKey="date"
          tick={{ fill: '#555', fontSize: 9 }}
          axisLine={false}
          tickLine={false}
          interval="preserveStartEnd"
        />
        <YAxis
          tick={{ fill: '#555', fontSize: 9 }}
          axisLine={false}
          tickLine={false}
          tickFormatter={formatK}
          width={32}
        />
        <Tooltip
          contentStyle={{ background: '#0f0f0f', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 11 }}
          labelStyle={{ color: '#888' }}
          itemStyle={{ color: '#d97706' }}
          formatter={(v: number) => [formatK(v), 'tokens']}
        />
        <Area
          type="monotone"
          dataKey="tokens"
          stroke="#d97706"
          strokeWidth={1.5}
          fill="url(#tokenGrad)"
          dot={false}
          activeDot={{ r: 3, fill: '#d97706' }}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
