import ReactMarkdown from 'react-markdown'
import { PrismLight as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'

import typescript from 'react-syntax-highlighter/dist/esm/languages/prism/typescript'
import tsx from 'react-syntax-highlighter/dist/esm/languages/prism/tsx'
import javascript from 'react-syntax-highlighter/dist/esm/languages/prism/javascript'
import jsx from 'react-syntax-highlighter/dist/esm/languages/prism/jsx'
import python from 'react-syntax-highlighter/dist/esm/languages/prism/python'
import bash from 'react-syntax-highlighter/dist/esm/languages/prism/bash'
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json'
import yaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml'
import go from 'react-syntax-highlighter/dist/esm/languages/prism/go'
import sql from 'react-syntax-highlighter/dist/esm/languages/prism/sql'
import css from 'react-syntax-highlighter/dist/esm/languages/prism/css'
import markup from 'react-syntax-highlighter/dist/esm/languages/prism/markup'
import markdown from 'react-syntax-highlighter/dist/esm/languages/prism/markdown'
import powershell from 'react-syntax-highlighter/dist/esm/languages/prism/powershell'
import csharp from 'react-syntax-highlighter/dist/esm/languages/prism/csharp'

SyntaxHighlighter.registerLanguage('typescript', typescript)
SyntaxHighlighter.registerLanguage('tsx', tsx)
SyntaxHighlighter.registerLanguage('javascript', javascript)
SyntaxHighlighter.registerLanguage('js', javascript)
SyntaxHighlighter.registerLanguage('jsx', jsx)
SyntaxHighlighter.registerLanguage('python', python)
SyntaxHighlighter.registerLanguage('py', python)
SyntaxHighlighter.registerLanguage('bash', bash)
SyntaxHighlighter.registerLanguage('sh', bash)
SyntaxHighlighter.registerLanguage('shell', bash)
SyntaxHighlighter.registerLanguage('json', json)
SyntaxHighlighter.registerLanguage('yaml', yaml)
SyntaxHighlighter.registerLanguage('yml', yaml)
SyntaxHighlighter.registerLanguage('go', go)
SyntaxHighlighter.registerLanguage('sql', sql)
SyntaxHighlighter.registerLanguage('css', css)
SyntaxHighlighter.registerLanguage('html', markup)
SyntaxHighlighter.registerLanguage('xml', markup)
SyntaxHighlighter.registerLanguage('markdown', markdown)
SyntaxHighlighter.registerLanguage('md', markdown)
SyntaxHighlighter.registerLanguage('powershell', powershell)
SyntaxHighlighter.registerLanguage('ps1', powershell)
SyntaxHighlighter.registerLanguage('csharp', csharp)
SyntaxHighlighter.registerLanguage('cs', csharp)

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
          className="text-text-primary text-xs leading-relaxed prose-invert"
          components={{
            code({ className, children, ...props }) {
              const langMatch = (className ?? '').match(/language-(\w+)/)
              const isInline = !langMatch
              return isInline ? (
                <code className="bg-bg-elevated text-accent px-1 py-0.5 rounded text-[10px]" {...props}>
                  {children}
                </code>
              ) : (
                <SyntaxHighlighter
                  style={vscDarkPlus}
                  language={langMatch[1]}
                  PreTag="div"
                  customStyle={{ margin: '8px 0', borderRadius: 4, fontSize: 10 }}
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
