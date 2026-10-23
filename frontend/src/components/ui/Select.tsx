import * as RS from '@radix-ui/react-select'
import { Check, ChevronDown } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

export interface Option<T extends string> {
  value: T
  label: ReactNode
  hint?: ReactNode
  /** Text direction for this option (e.g. native language names). */
  dir?: 'ltr' | 'rtl' | 'auto'
}

interface SelectProps<T extends string> {
  value: T
  onChange(v: T): void
  options: Option<T>[]
  label?: string
  className?: string
  disabled?: boolean
  size?: 'sm' | 'md'
}

// Radix Select doesn't allow "" as an item value; map it to a sentinel.
const EMPTY = '__empty__'
const enc = (v: string) => (v === '' ? EMPTY : v)
const dec = (v: string) => (v === EMPTY ? '' : v)

export function Select<T extends string>({ value, onChange, options, label, className, disabled, size = 'md' }: SelectProps<T>) {
  const current = options.find((o) => o.value === value)
  return (
    <RS.Root value={enc(value)} onValueChange={(v) => onChange(dec(v) as T)} disabled={disabled} dir={document.documentElement.dir as 'ltr' | 'rtl'}>
      <RS.Trigger
        aria-label={label}
        className={cn(
          'no-drag inline-flex min-w-0 items-center justify-between gap-2 rounded-xl border bg-elevated text-start text-fg transition-colors',
          'hover:bg-line/[0.04] focus-visible:shadow-focus disabled:opacity-50 dark:bg-line/[0.04]',
          size === 'sm' ? 'h-8 px-2.5 text-xs' : 'h-9 px-3 text-[13px]',
          className,
        )}
      >
        <span className="truncate" dir={current?.dir}>
          <RS.Value>{current?.label}</RS.Value>
        </span>
        <RS.Icon>
          <ChevronDown className="size-4 shrink-0 text-subtle" aria-hidden />
        </RS.Icon>
      </RS.Trigger>
      <RS.Portal>
        <RS.Content
          position="popper"
          sideOffset={6}
          collisionPadding={8}
          className="z-[90] max-h-[min(360px,var(--radix-select-content-available-height))] min-w-[var(--radix-select-trigger-width)] overflow-hidden rounded-xl border bg-elevated p-1 shadow-lift"
        >
          <RS.Viewport>
            {options.map((o) => (
              <RS.Item
                key={o.value || EMPTY}
                value={enc(o.value)}
                className="relative flex cursor-default select-none items-center gap-2 rounded-lg py-2 pe-8 ps-2.5 text-[13px] text-fg outline-none data-[highlighted]:bg-accent/10 data-[disabled]:opacity-40"
              >
                <div className="min-w-0">
                  <RS.ItemText>
                    <span dir={o.dir}>{o.label}</span>
                  </RS.ItemText>
                  {o.hint && <div className="truncate text-xs text-subtle">{o.hint}</div>}
                </div>
                <RS.ItemIndicator className="absolute end-2.5">
                  <Check className="size-4 text-accent-ink" aria-hidden />
                </RS.ItemIndicator>
              </RS.Item>
            ))}
          </RS.Viewport>
        </RS.Content>
      </RS.Portal>
    </RS.Root>
  )
}
