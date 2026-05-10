import { useState, useRef, useCallback } from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { useSession } from '../context/SessionContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { useDashboard } from '../hooks/useDashboard'
import type { Session } from '../hooks/useDashboard'
import StatsStrip from '../components/StatsStrip'
import SessionList from '../components/SessionList'
import SessionHeader from '../components/SessionHeader'
import MessageList from '../components/MessageList'
import ChatInput from '../components/ChatInput'
import TerminalDrawer, { type TerminalDrawerHandle } from '../components/TerminalDrawer'
import NewSessionModal from '../components/NewSessionModal'
import UsageChart from '../components/UsageChart'

export default function DashboardView() {
  const { state, dispatch } = useSession()
  const terminalRef = useRef<TerminalDrawerHandle>(null)
  const [showModal, setShowModal] = useState(false)
  const [chartOpen, setChartOpen] = useState(false)

  const { account, usage, sessions, activeSessions, loading, refresh } = useDashboard()

  // Local sessions state for optimistic deletion
  const [localSessions, setLocalSessions] = useState<Session[] | null>(null)
  const displaySessions = localSessions ?? sessions

  // Keep local sessions in sync when sessions fetched from API
  const prevSessionsRef = useRef(sessions)
  if (prevSessionsRef.current !== sessions) {
    prevSessionsRef.current = sessions
    setLocalSessions(null) // reset to API data on each refresh
  }

  const onOutput = useCallback((data: string) => {
    terminalRef.current?.write(data)
  }, [])

  const { send } = useWebSocket(onOutput)

  // Wire terminal resize → backend PTY resize
  const onTerminalMounted = useCallback(() => {
    terminalRef.current?.sendResize((cols, rows) => {
      send({ type: 'resize', cols, rows })
    })
  }, [send])
  void onTerminalMounted

  function handleSessionStart(sessionId: string, workdir: string) {
    dispatch({ type: 'SESSION_CREATED', sessionId, workdir })
    if (account?.model) dispatch({ type: 'MODEL_SET', model: account.model })
    setShowModal(false)
    refresh()
  }

  function handleNewSession() {
    if (state.sessionId) {
      fetch(`/api/sessions/${state.sessionId}/stop`, { method: 'POST' }).catch(() => {})
    }
    dispatch({ type: 'SESSION_CLEARED' })
    setShowModal(true)
  }

  function handleSessionsList() {
    // Detach from the session (PTY keeps running on backend; user can reconnect from the list)
    dispatch({ type: 'SESSION_CLEARED' })
  }

  function handleSend(text: string) {
    dispatch({
      type: 'MESSAGE_ADDED',
      message: { id: Date.now().toString(36) + Math.random().toString(36).slice(2), role: 'user', content: text, createdAt: Date.now() },
    })
    // Use type:'chat' for structured chat messages (terminal raw input uses type:'input')
    send({ type: 'chat', data: text + '\n' })
  }

  function handleResume(session: Session) {
    dispatch({ type: 'RESUME_SESSION', id: session.id, workdir: session.workdir })
    if (account?.model) dispatch({ type: 'MODEL_SET', model: account.model })
  }

  function handleDelete(sessionId: string) {
    setLocalSessions((prev) => {
      const base = prev ?? sessions
      return base.filter((s) => s.id !== sessionId)
    })
  }

  // Find active session for the live duration timer
  const activeSession = state.sessionId
    ? displaySessions.find((s) => s.id === state.sessionId && s.is_active) ?? null
    : null

  return (
    <div className="flex flex-col h-full">
      {/* Stats strip — account/general stats only */}
      <StatsStrip account={account} usage={usage} activeSessions={activeSessions} loading={loading} />

      {/* Collapsible usage chart */}
      <div className="border-b border-border-subtle flex-shrink-0">
        <button
          onClick={() => setChartOpen((o) => !o)}
          className="w-full px-4 py-1.5 flex items-center gap-2 hover:bg-bg-elevated transition-colors"
        >
          {chartOpen
            ? <ChevronUp size={12} className="text-text-dim" />
            : <ChevronDown size={12} className="text-text-dim" />}
          <span className="text-text-dim text-xs uppercase tracking-widest">Usage</span>
        </button>
        {chartOpen && (
          <div className="px-2 pb-2">
            <UsageChart days={usage?.days ?? []} />
          </div>
        )}
      </div>

      {/* Main area */}
      {state.sessionId ? (
        <div className="flex flex-col flex-1 overflow-hidden">
          <SessionHeader
            onNewSession={handleNewSession}
            onSessionsList={handleSessionsList}
            totalTokens={(usage?.totals.inputTokens ?? 0) + (usage?.totals.outputTokens ?? 0)}
            sessionStartedAt={activeSession?.started_at ?? null}
          />
          <MessageList messages={state.messages} />
          <TerminalDrawer ref={terminalRef} wsState={state.wsState} />
          <ChatInput
            onSend={handleSend}
            disabled={state.wsState === 'disconnected' || state.wsState === 'error'}
          />
        </div>
      ) : showModal ? (
        <NewSessionModal onStart={handleSessionStart} />
      ) : (
        <SessionList
          sessions={displaySessions}
          onResume={handleResume}
          onDelete={handleDelete}
          onNewSession={() => setShowModal(true)}
        />
      )}
    </div>
  )
}
