import * as RT from '@radix-ui/react-tooltip'
import type { ReactNode } from 'react'

export function TooltipProvider({ children }: { children: ReactNode }) {
  return (
    <RT.Provider delayDuration={350} skipDelayDuration={200}>
      {children}
    </RT.Provider>
  )
}

export function Tooltip({ content, children, side = 'top' }: { content: ReactNode; children: ReactNode; side?: 'top' | 'bottom' | 'left' | 'right' }) {
  if (!content) return <>{children}</>
  return (
    <RT.Root>
      <RT.Trigger asChild>{children}</RT.Trigger>
      <RT.Portal>
        <RT.Content
          side={side}
          sideOffset={6}
          collisionPadding={8}
          className="z-[100] max-w-xs rounded-lg border bg-elevated px-2.5 py-1.5 text-xs text-fg shadow-lift data-[state=delayed-open]:animate-[fadein_120ms_ease-out]"
        >
          {content}
        </RT.Content>
      </RT.Portal>
    </RT.Root>
  )
}
