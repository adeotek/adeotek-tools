import { createContext, useContext, useReducer, type ReactNode } from 'react'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: number
}

export interface SessionState {
  sessionId: string | null
  workdir: string | null
  name: string | null
  model: string | null
  wsState: 'disconnected' | 'connecting' | 'running' | 'idle' | 'error'
  messages: Message[]
  workingTimeMs: number       // cumulative ms spent in 'running' state
  runningStartedAt: number | null  // timestamp when current run period began
}

type Action =
  | { type: 'SESSION_CREATED'; sessionId: string; workdir: string; name?: string }
  | { type: 'SESSION_CLEARED' }
  | { type: 'RESUME_SESSION'; id: string; workdir: string; name?: string }
  | { type: 'WS_STATE'; state: SessionState['wsState']; timestamp: number }
  | { type: 'MESSAGE_ADDED'; message: Message }
  | { type: 'HISTORY_LOADED'; messages: Message[] }
  | { type: 'MODEL_SET'; model: string }
  | { type: 'SESSION_RENAMED'; name: string | null }

const initial: SessionState = {
  sessionId: null,
  workdir: null,
  name: null,
  model: null,
  wsState: 'disconnected',
  messages: [],
  workingTimeMs: 0,
  runningStartedAt: null,
}

function reducer(state: SessionState, action: Action): SessionState {
  switch (action.type) {
    case 'SESSION_CREATED':
      return { ...state, sessionId: action.sessionId, workdir: action.workdir, name: action.name ?? null, messages: [], wsState: 'connecting', workingTimeMs: 0, runningStartedAt: null }
    case 'SESSION_CLEARED':
      return { ...initial }
    case 'RESUME_SESSION':
      return { ...state, sessionId: action.id, workdir: action.workdir, name: action.name ?? null, messages: [], wsState: 'connecting', workingTimeMs: 0, runningStartedAt: null }
    case 'WS_STATE': {
      const prev = state.wsState
      const next = action.state
      let { workingTimeMs, runningStartedAt } = state
      if (next === 'running' && prev !== 'running') {
        runningStartedAt = action.timestamp
      } else if (prev === 'running' && next !== 'running' && runningStartedAt != null) {
        workingTimeMs += action.timestamp - runningStartedAt
        runningStartedAt = null
      }
      return { ...state, wsState: next, workingTimeMs, runningStartedAt }
    }
    case 'MESSAGE_ADDED':
      return { ...state, messages: [...state.messages, action.message] }
    case 'HISTORY_LOADED':
      return { ...state, messages: [...action.messages, ...state.messages] }
    case 'MODEL_SET':
      return { ...state, model: action.model }
    case 'SESSION_RENAMED':
      return { ...state, name: action.name }
    default:
      return state
  }
}

const SessionContext = createContext<{
  state: SessionState
  dispatch: React.Dispatch<Action>
} | null>(null)

export function SessionProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initial)
  return <SessionContext.Provider value={{ state, dispatch }}>{children}</SessionContext.Provider>
}

export function useSession() {
  const ctx = useContext(SessionContext)
  if (!ctx) throw new Error('useSession must be used inside SessionProvider')
  return ctx
}
