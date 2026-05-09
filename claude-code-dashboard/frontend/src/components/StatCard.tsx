interface StatCardProps {
  label: string
  value: string
  color?: 'amber' | 'green' | 'blue'
}

const colorMap = {
  amber: 'text-accent',
  green: 'text-status-green',
  blue: 'text-status-blue',
}

export default function StatCard({ label, value, color = 'amber' }: StatCardProps) {
  return (
    <div className="bg-bg-surface border border-border-subtle rounded-md p-4 text-center">
      <div className={`text-lg font-semibold ${colorMap[color]}`}>{value}</div>
      <div className="text-text-muted text-xs mt-1 uppercase tracking-widest">{label}</div>
    </div>
  )
}
