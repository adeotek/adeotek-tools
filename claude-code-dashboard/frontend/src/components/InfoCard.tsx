interface InfoCardProps {
  label: string
  value: string | null
  sub?: string
  accent?: boolean
}

export default function InfoCard({ label, value, sub, accent }: InfoCardProps) {
  return (
    <div className="bg-bg-surface border border-border-subtle rounded-md p-4">
      <div className="text-text-muted text-xs uppercase tracking-widest mb-2">{label}</div>
      <div className={`text-sm font-medium ${accent ? 'text-accent' : 'text-text-primary'}`}>
        {value ?? <span className="text-text-dim">—</span>}
      </div>
      {sub && <div className="text-text-muted text-xs mt-1">{sub}</div>}
    </div>
  )
}
