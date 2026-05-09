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

    function connect() {
      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const host = window.location.host
      const ws = new WebSocket(`${protocol}://${host}/ws/session/${state.sessionId}`)
      wsRef.current = ws
      dispatch({ type: 'WS_STATE', state: 'connecting' })

      ws.onopen = () => {
        attemptsRef.current = 0
        dispatch({ type: 'WS_STATE', state: 'running' })
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data as string) as {
            type: string; data?: string; role?: string; content?: string; state?: string
          }
          if (msg.type === 'output' && msg.data) {
            onOutput(msg.data)
          } else if (msg.type === 'message' && msg.role === 'assistant' && msg.content) {
            dispatch({
              type: 'MESSAGE_ADDED',
              message: { id: crypto.randomUUID(), role: 'assistant', content: msg.content, createdAt: Date.now() },
            })
          } else if (msg.type === 'status' && msg.state) {
            dispatch({ type: 'WS_STATE', state: msg.state as 'running' | 'idle' | 'error' })
          }
        } catch {
          // ignore malformed frames
        }
      }

      ws.onerror = () => dispatch({ type: 'WS_STATE', state: 'error' })

      ws.onclose = () => {
        if (attemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          const delay = BASE_DELAY_MS * 2 ** attemptsRef.current
          attemptsRef.current++
          setTimeout(connect, delay)
        } else {
          dispatch({ type: 'WS_STATE', state: 'disconnected' })
        }
      }
    }

    connect()
    return () => {
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
