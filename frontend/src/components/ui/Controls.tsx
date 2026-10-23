import * as RSw from '@radix-ui/react-switch'
import * as RC from '@radix-ui/react-checkbox'
import { Check, Minus } from 'lucide-react'
import { motion } from 'framer-motion'
import { forwardRef, useId, type InputHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

export function Switch({ checked, onChange, label, disabled }: { checked: boolean; onChange(v: boolean): void; label: string; disabled?: boolean }) {
  return (
    <RSw.Root
      checked={checked}
      onCheckedChange={onChange}
      disabled={disabled}
      aria-label={label}
      className={cn(
        'no-drag relative inline-flex h-[22px] w-[38px] shrink-0 items-center rounded-full border border-transparent transition-colors duration-200',
        'bg-line/15 data-[state=checked]:bg-accent focus-visible:shadow-focus disabled:opacity-50',
      )}
    >
      <RSw.Thumb
        className={cn(
          'block size-[18px] rounded-full bg-white shadow-sm transition-transform duration-200',
          'translate-x-[1px] data-[state=checked]:translate-x-[17px]',
          'rtl:-translate-x-[1px] rtl:data-[state=checked]:-translate-x-[17px]',
        )}
      />
    </RSw.Root>
  )
}

export function Checkbox({ checked, onChange, label, indeterminate }: { checked: boolean; onChange(v: boolean): void; label: string; indeterminate?: boolean }) {
  return (
    <RC.Root
      checked={indeterminate ? 'indeterminate' : checked}
      onCheckedChange={(v) => onChange(v === true)}
      aria-label={label}
      className="no-drag flex size-[18px] shrink-0 items-center justify-center rounded-[6px] border border-line/25 bg-transparent transition-colors data-[state=checked]:border-accent data-[state=indeterminate]:border-accent data-[state=checked]:bg-accent data-[state=indeterminate]:bg-accent focus-visible:shadow-focus"
    >
      <RC.Indicator className="text-accent-fg">
        {indeterminate ? <Minus className="size-3" strokeWidth={3} /> : <Check className="size-3" strokeWidth={3} />}
      </RC.Indicator>
    </RC.Root>
  )
}

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement> & { icon?: ReactNode }>(function Input(
  { className, icon, ...rest },
  ref,
) {
  return (
    <div className={cn('relative flex items-center', className)}>
      {icon && <span className="pointer-events-none absolute start-3 text-subtle">{icon}</span>}
      <input
        ref={ref}
        className={cn(
          'no-drag h-9 w-full rounded-xl border bg-elevated px-3 text-[13px] text-fg placeholder:text-subtle transition-shadow dark:bg-line/[0.04]',
          'focus-visible:border-accent/50 focus-visible:shadow-focus focus-visible:outline-none',
          icon && 'ps-9',
        )}
        {...rest}
      />
    </div>
  )
})

export interface SegmentOption<T extends string> {
  value: T
  label: ReactNode
  disabled?: boolean
  title?: string
}

/** Accessible radio group rendered as a segmented control with an animated pill. */
export function Segmented<T extends string>({
  value,
  onChange,
  options,
  label,
  size = 'md',
  className,
}: {
  value: T
  onChange(v: T): void
  options: SegmentOption<T>[]
  label: string
  size?: 'sm' | 'md'
  className?: string
}) {
  const id = useId()
  const move = (dir: number) => {
    const enabled = options.filter((o) => !o.disabled)
    const i = enabled.findIndex((o) => o.value === value)
    const rtl = document.documentElement.dir === 'rtl' ? -1 : 1
    const next = enabled[(i + dir * rtl + enabled.length) % enabled.length]
    if (next) onChange(next.value)
  }
  return (
    <div
      role="radiogroup"
      aria-label={label}
      className={cn('no-drag inline-flex rounded-xl border bg-line/[0.04] p-0.5', className)}
      onKeyDown={(e) => {
        if (e.key === 'ArrowRight' || e.key === 'ArrowDown') (e.preventDefault(), move(1))
        if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') (e.preventDefault(), move(-1))
      }}
    >
      {options.map((o) => {
        const active = o.value === value
        return (
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={active}
            tabIndex={active ? 0 : -1}
            disabled={o.disabled}
            title={o.title}
            onClick={() => onChange(o.value)}
            className={cn(
              'relative flex-1 whitespace-nowrap rounded-[10px] font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-35',
              size === 'sm' ? 'h-7 px-2.5 text-xs' : 'h-8 px-3 text-[13px]',
              active ? 'text-fg' : 'text-muted hover:text-fg',
            )}
          >
            {active && (
              <motion.span
                layoutId={`seg-${id}`}
                className="absolute inset-0 rounded-[10px] border bg-elevated shadow-sm dark:bg-line/[0.09]"
                transition={{ type: 'spring', stiffness: 500, damping: 38 }}
              />
            )}
            <span className="relative">{o.label}</span>
          </button>
        )
      })}
    </div>
  )
}

export function Badge({ children, tone = 'neutral', className, title }: { children: ReactNode; tone?: 'neutral' | 'accent' | 'success' | 'warning' | 'danger'; className?: string; title?: string }) {
  const tones = {
    neutral: 'bg-line/[0.07] text-muted',
    accent: 'bg-accent/[0.12] text-accent-ink',
    success: 'bg-success/[0.12] text-success',
    warning: 'bg-warning/[0.12] text-warning',
    danger: 'bg-danger/[0.12] text-danger',
  }
  return (
    <span title={title} className={cn('inline-flex h-5 shrink-0 items-center gap-1 rounded-md px-1.5 text-2xs font-semibold', tones[tone], className)}>
      {children}
    </span>
  )
}

export function Kbd({ children }: { children: ReactNode }) {
  return (
    <kbd dir="ltr" className="inline-flex h-5 min-w-5 items-center justify-center rounded-md border border-b-2 bg-elevated px-1.5 font-sans text-2xs font-medium text-muted">
      {children}
    </kbd>
  )
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('relative overflow-hidden rounded-lg bg-line/[0.07]', className)}><Shimmer /></div>
}

export function Shimmer({ className }: { className?: string }) {
  return (
    <span
      aria-hidden
      className={cn('pointer-events-none absolute inset-y-0 w-1/2 animate-shimmer bg-gradient-to-r from-transparent via-white/[0.08] to-transparent', className)}
      style={{ insetInlineStart: 0 }}
    />
  )
}
