import { HashRouter, Routes, Route, Navigate } from 'react-router-dom'
import DashboardView from './views/DashboardView'
import SettingsView from './views/SettingsView'
import { SessionProvider } from './context/SessionContext'

export default function App() {
  return (
    <HashRouter>
      <SessionProvider>
        <div className="flex h-screen bg-bg-base text-text-primary overflow-hidden font-mono">
          <main className="flex-1 overflow-hidden flex flex-col">
            <Routes>
              <Route path="/" element={<DashboardView />} />
              <Route path="/settings" element={<SettingsView />} />
              <Route path="/account" element={<Navigate to="/" replace />} />
              <Route path="/usage" element={<Navigate to="/" replace />} />
              <Route path="/chat" element={<Navigate to="/" replace />} />
            </Routes>
          </main>
        </div>
      </SessionProvider>
    </HashRouter>
  )
}
