import { HashRouter, Routes, Route, Navigate } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import AccountView from './views/AccountView'
import UsageView from './views/UsageView'
import ChatView from './views/ChatView'
import SettingsView from './views/SettingsView'
import { SessionProvider } from './context/SessionContext'

export default function App() {
  return (
    <HashRouter>
      <SessionProvider>
        <div className="flex h-screen bg-bg-base text-text-primary overflow-hidden font-mono">
          <Sidebar />
          <main className="flex-1 overflow-hidden flex flex-col">
            <Routes>
              <Route path="/" element={<Navigate to="/chat" replace />} />
              <Route path="/account" element={<AccountView />} />
              <Route path="/usage" element={<UsageView />} />
              <Route path="/chat" element={<ChatView />} />
              <Route path="/settings" element={<SettingsView />} />
            </Routes>
          </main>
        </div>
      </SessionProvider>
    </HashRouter>
  )
}
