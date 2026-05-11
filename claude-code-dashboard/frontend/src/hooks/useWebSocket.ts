import { useEffect, useRef, useCallback } from 'react'
import { useSession } from '../context/SessionContext'

const MAX_RECONNECT_ATTEMPTS = 5
const BASE_DELAY_MS = 500

export function useWebSocket(onOutput: (data: string) => void) {
  const { state, dispatch } = useSession()
  const wsRef = useRef<WebSocket | null>(null)
  const attemptsRef = useRef(0)

  useEffect(() => {
    if (!state.sessionId) return
    let closed = false

    function connect() {
      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const host = window.location.host
      const ws = new WebSocket(`${protocol}://${host}/ws/session/${state.sessionId}`)
      wsRef.current = ws
      dispatch({ type: 'WS_STATE', timestamp: Date.now(), state: 'connecting' })

      ws.onopen = () => {
        attemptsRef.current = 0
        dispatch({ type: 'WS_STATE', timestamp: Date.now(), state: 'idle' })
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data as string) as {
            type: string
            data?: string
            role?: string
            content?: string
            state?: string
            messages?: Array<{ role: string; content: string; created_at: number }>
            inputTokens?: number
            outputTokens?: number
            totalTokens?: number
            workingTimeMs?: number
          }
          if (msg.type === 'output' && msg.data) {
            onOutput(msg.data)
          } else if (msg.type === 'message' && msg.role === 'assistant' && msg.content) {
            dispatch({
              type: 'MESSAGE_ADDED',
              message: { id: Date.now().toString(36) + Math.random().toString(36).slice(2), role: 'assistant', content: msg.content, createdAt: Date.now() },
            })
          } else if (msg.type === 'status' && msg.state) {
            dispatch({ type: 'WS_STATE', timestamp: Date.now(), state: msg.state as 'running' | 'idle' | 'error' })
          } else if (msg.type === 'tokens' && msg.inputTokens != null && msg.outputTokens != null) {
            dispatch({ type: 'TOKENS_ADDED', inputTokens: msg.inputTokens, outputTokens: msg.outputTokens })
          } else if (msg.type === 'session_state') {
            dispatch({ type: 'STATS_RESTORED', totalTokens: msg.totalTokens ?? 0, workingTimeMs: msg.workingTimeMs ?? 0 })
          } else if (msg.type === 'history' && Array.isArray(msg.messages)) {
            const history = msg.messages.map((m) => ({
              id: Date.now().toString(36) + Math.random().toString(36).slice(2),
              role: m.role as 'user' | 'assistant',
              content: m.content,
              createdAt: m.created_at,
            }))
            dispatch({ type: 'HISTORY_LOADED', messages: history })
          }
        } catch {
          // ignore malformed frames
        }
      }

      ws.onerror = () => dispatch({ type: 'WS_STATE', timestamp: Date.now(), state: 'error' })

      ws.onclose = () => {
        if (closed) return
        if (attemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          const delay = BASE_DELAY_MS * 2 ** attemptsRef.current
          attemptsRef.current++
          setTimeout(connect, delay)
        } else {
          dispatch({ type: 'WS_STATE', timestamp: Date.now(), state: 'disconnected' })
          // Keep a slow background retry so the connection self-heals when the server comes back
          setTimeout(() => { if (!closed) { attemptsRef.current = 0; connect() } }, 30_000)
        }
      }
    }

    connect()
    return () => {
      closed = true
      attemptsRef.current = 0
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [state.sessionId]) // eslint-disable-line react-hooks/exhaustive-deps

  const send = useCallback((payload: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(payload))
    }
  }, [])

  return { send }
}
