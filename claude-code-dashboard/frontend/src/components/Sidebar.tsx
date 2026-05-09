import { NavLink } from 'react-router-dom'
import { User, BarChart2, MessageSquare, Settings } from 'lucide-react'

const navItems = [
  { to: '/account', icon: User, label: 'Account' },
  { to: '/usage', icon: BarChart2, label: 'Usage' },
  { to: '/chat', icon: MessageSquare, label: 'Chat' },
]

export default function Sidebar() {
  return (
    <nav className="w-12 bg-bg-surface border-r border-border-subtle flex flex-col items-center py-3 gap-1.5 flex-shrink-0">
      <div className="w-7 h-7 rounded-md bg-accent flex items-center justify-center mb-3 flex-shrink-0 text-black font-bold text-xs">
        ⚡
      </div>

      {navItems.map(({ to, icon: Icon, label }) => (
        <NavLink
          key={to}
          to={to}
          title={label}
          className={({ isActive }) =>
            'w-9 h-9 rounded-md flex items-center justify-center transition-colors ' +
            (isActive
              ? 'bg-accent text-black'
              : 'text-text-muted hover:text-text-secondary hover:bg-bg-elevated')
          }
        >
          <Icon size={16} strokeWidth={1.5} />
        </NavLink>
      ))}

      <div className="mt-auto">
        <NavLink
          to="/settings"
          title="Settings"
          className={({ isActive }) =>
            'w-9 h-9 rounded-md flex items-center justify-center transition-colors ' +
            (isActive
              ? 'bg-accent text-black'
              : 'text-text-dim hover:text-text-muted hover:bg-bg-elevated')
          }
        >
          <Settings size={16} strokeWidth={1.5} />
        </NavLink>
      </div>
    </nav>
  )
}
