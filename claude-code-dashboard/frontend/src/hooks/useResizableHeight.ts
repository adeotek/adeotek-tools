import { useState, useRef, useCallback } from 'react'

export function useResizableHeight(initial: number, min: number, max: number) {
  const [height, setHeight] = useState(initial)
  const heightRef = useRef(initial)

  const onDragStart = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      const startY = e.clientY
      const startH = heightRef.current

      document.body.style.cursor = 'ns-resize'
      document.body.style.userSelect = 'none'

      const onMove = (ev: MouseEvent) => {
        const newH = Math.max(min, Math.min(max, startH + (startY - ev.clientY)))
        heightRef.current = newH
        setHeight(newH)
      }

      const onUp = () => {
        document.body.style.cursor = ''
        document.body.style.userSelect = ''
        document.removeEventListener('mousemove', onMove)
        document.removeEventListener('mouseup', onUp)
      }

      document.addEventListener('mousemove', onMove)
      document.addEventListener('mouseup', onUp)
    },
    [min, max],
  )

  return { height, onDragStart }
}
