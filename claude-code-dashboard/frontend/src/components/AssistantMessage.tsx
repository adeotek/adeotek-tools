import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface AssistantMessageProps {
  content: string
}

export default function AssistantMessage({ content }: AssistantMessageProps) {
  return (
    <div className="flex gap-2.5">
      <div className="w-6 h-6 rounded-full bg-bg-panel border border-border flex-shrink-0 mt-0.5 flex items-center justify-center text-status-blue text-xs font-bold">
        C
      </div>
      <div className="max-w-[85%] bg-bg-surface border border-border-subtle rounded-lg rounded-bl-sm px-3 py-2">
        <ReactMarkdown
          className="text-text-primary text-sm leading-relaxed prose-invert"
          components={{
            code({ className, children, ...props }) {
              const langMatch = (className ?? '').match(/language-(\w+)/)
              const isInline = !langMatch
              return isInline ? (
                <code className="bg-bg-elevated text-accent px-1 py-0.5 rounded text-xs" {...props}>
                  {children}
                </code>
              ) : (
                <SyntaxHighlighter
                  style={vscDarkPlus}
                  language={langMatch[1]}
                  PreTag="div"
                  customStyle={{ margin: '8px 0', borderRadius: 4, fontSize: 11 }}
                >
                  {String(children).replace(/\n$/, '')}
                </SyntaxHighlighter>
              )
            },
            p({ children }) {
              return <p className="mb-2 last:mb-0">{children}</p>
            },
          }}
        >
          {content}
        </ReactMarkdown>
      </div>
    </div>
  )
}
