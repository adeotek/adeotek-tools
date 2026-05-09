import { useEffect, useRef, useState, useImperativeHandle, forwardRef } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { ChevronUp, ChevronDown } from 'lucide-react'

export interface TerminalDrawerHandle {
  write: (data: string) => void
  sendResize: (cb: (cols: number, rows: number) => void) => void
}

const TerminalDrawer = forwardRef<TerminalDrawerHandle, { wsState: string }>(
  ({ wsState }, ref) => {
    const [open, setOpen] = useState(false)
    const containerRef = useRef<HTMLDivElement>(null)
    const termRef = useRef<Terminal | null>(null)
    const fitRef = useRef<FitAddon | null>(null)
    const onResizeCbRef = useRef<((cols: number, rows: number) => void) | null>(null)

    useEffect(() => {
      const term = new Terminal({
        theme: {
          background: '#050505',
          foreground: '#e2e2e2',
          cursor: '#d97706',
          selectionBackground: '#d9770640',
        },
        fontFamily: 'JetBrains Mono, Fira Code, monospace',
        fontSize: 12,
        lineHeight: 1.4,
        cursorBlink: true,
      })
      const fit = new FitAddon()
      term.loadAddon(fit)
      termRef.current = term
      fitRef.current = fit

      if (containerRef.current) {
        term.open(containerRef.current)
        fit.fit()
        term.onResize(({ cols, rows }) => onResizeCbRef.current?.(cols, rows))
      }

      return () => term.dispose()
    }, [])

    useEffect(() => {
      if (open) setTimeout(() => fitRef.current?.fit(), 10)
    }, [open])

    useImperativeHandle(ref, () => ({
      write: (data) => termRef.current?.write(data),
      sendResize: (cb) => { onResizeCbRef.current = cb },
    }))

    const isRunning = wsState === 'running'

    return (
      <div className="border-t border-border-subtle bg-bg-base flex-shrink-0">
        <button
          onClick={() => setOpen((o) => !o)}
          className="w-full px-4 py-1.5 flex items-center gap-2 hover:bg-bg-elevated transition-colors"
        >
          {open ? <ChevronDown size={12} className="text-text-dim" /> : <ChevronUp size={12} className="text-text-dim" />}
          <span className="text-text-dim text-xs uppercase tracking-widest">Terminal output</span>
          <div className="ml-auto flex items-center gap-1.5">
            <div className={`w-1.5 h-1.5 rounded-full ${isRunning ? 'bg-status-green' : 'bg-text-dim'}`} />
            <span className="text-text-dim text-xs">{wsState}</span>
          </div>
        </button>

        {/* xterm is always mounted, just hidden — preserves PTY scroll buffer */}
        <div style={{ display: open ? 'block' : 'none', height: 200 }}>
          <div ref={containerRef} className="h-full" />
        </div>
      </div>
    )
  },
)

TerminalDrawer.displayName = 'TerminalDrawer'
export default TerminalDrawer
