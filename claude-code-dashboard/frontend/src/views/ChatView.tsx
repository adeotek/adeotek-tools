import { useRef, useCallback } from 'react'
import { AlertTriangle } from 'lucide-react'
import { useSession } from '../context/SessionContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAccount } from '../hooks/useAccount'
import MessageList from '../components/MessageList'
import ChatInput from '../components/ChatInput'
import SessionHeader from '../components/SessionHeader'
import NewSessionModal from '../components/NewSessionModal'
import TerminalDrawer, { type TerminalDrawerHandle } from '../components/TerminalDrawer'

export default function ChatView() {
  const { state, dispatch } = useSession()
  const { data: account } = useAccount()
  const terminalRef = useRef<TerminalDrawerHandle>(null)

  const onOutput = useCallback((data: string) => {
    terminalRef.current?.write(data)
  }, [])

  const { send } = useWebSocket(onOutput)

  // Wire terminal resize → backend PTY resize
  const terminalDrawerElement = (
    <TerminalDrawer
      ref={terminalRef}
      wsState={state.wsState}
    />
  )

  // Once terminal mounts, register resize callback
  const onTerminalMounted = useCallback(() => {
    terminalRef.current?.sendResize((cols, rows) => {
      send({ type: 'resize', cols, rows })
    })
  }, [send])
  void onTerminalMounted // registered lazily by TerminalDrawer's useEffect

  function handleSessionStart(sessionId: string) {
    const workdir = (document.querySelector('input[type=text]') as HTMLInputElement | null)?.value ?? ''
    dispatch({ type: 'SESSION_CREATED', sessionId, workdir })
    if (account?.model) dispatch({ type: 'MODEL_SET', model: account.model })
  }

  function handleNewSession() {
    if (state.sessionId) {
      fetch(`/api/sessions/${state.sessionId}/stop`, { method: 'POST' }).catch(() => {})
    }
    dispatch({ type: 'SESSION_CLEARED' })
  }

  function handleSend(text: string) {
    dispatch({
      type: 'MESSAGE_ADDED',
      message: { id: crypto.randomUUID(), role: 'user', content: text, createdAt: Date.now() },
    })
    send({ type: 'input', data: text + '\n' })
  }

  if (account && !account.claudeInstalled) {
    return (
      <div className="flex-1 flex items-start gap-3 p-6">
        <AlertTriangle size={16} className="text-accent mt-0.5 flex-shrink-0" />
        <div>
          <p className="text-text-primary text-sm">Claude Code not found on this host</p>
          <p className="text-text-muted text-xs mt-1">
            Install and authenticate Claude Code, then reload the dashboard.
          </p>
        </div>
      </div>
    )
  }

  if (!state.sessionId) {
    return <NewSessionModal onStart={handleSessionStart} />
  }

  return (
    <div className="flex flex-col h-full">
      <SessionHeader onNewSession={handleNewSession} />
      <MessageList messages={state.messages} />
      {terminalDrawerElement}
      <ChatInput
        onSend={handleSend}
        disabled={state.wsState === 'disconnected' || state.wsState === 'error'}
      />
    </div>
  )
}
