interface UserMessageProps {
  content: string
}

export default function UserMessage({ content }: UserMessageProps) {
  return (
    <div className="flex gap-2.5 justify-end">
      <div className="max-w-[70%] bg-bg-elevated border border-border-subtle rounded-lg rounded-br-sm px-3 py-2">
        <p className="text-text-primary text-sm leading-relaxed whitespace-pre-wrap">{content}</p>
      </div>
      <div className="w-6 h-6 rounded-full bg-accent flex-shrink-0 mt-0.5 flex items-center justify-center text-black text-xs font-bold">
        U
      </div>
    </div>
  )
}
