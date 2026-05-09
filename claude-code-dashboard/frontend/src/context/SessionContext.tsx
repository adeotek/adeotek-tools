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
  model: string | null
  wsState: 'disconnected' | 'connecting' | 'running' | 'idle' | 'error'
  messages: Message[]
}

type Action =
  | { type: 'SESSION_CREATED'; sessionId: string; workdir: string }
  | { type: 'SESSION_CLEARED' }
  | { type: 'WS_STATE'; state: SessionState['wsState'] }
  | { type: 'MESSAGE_ADDED'; message: Message }
  | { type: 'MODEL_SET'; model: string }

const initial: SessionState = {
  sessionId: null,
  workdir: null,
  model: null,
  wsState: 'disconnected',
  messages: [],
}

function reducer(state: SessionState, action: Action): SessionState {
  switch (action.type) {
    case 'SESSION_CREATED':
      return { ...state, sessionId: action.sessionId, workdir: action.workdir, messages: [], wsState: 'connecting' }
    case 'SESSION_CLEARED':
      return { ...initial }
    case 'WS_STATE':
      return { ...state, wsState: action.state }
    case 'MESSAGE_ADDED':
      return { ...state, messages: [...state.messages, action.message] }
    case 'MODEL_SET':
      return { ...state, model: action.model }
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
