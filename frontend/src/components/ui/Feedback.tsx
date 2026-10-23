import * as RD from '@radix-ui/react-dialog'
import { AnimatePresence, motion } from 'framer-motion'
import { AlertTriangle, CheckCircle2, Info, X } from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useUI } from '@/store/ui'
import { Button } from './Button'
import { Shimmer } from './Controls'

/** Animated progress bar. Fills from the inline-start edge, so it mirrors in RTL. */
export function ProgressBar({
  value,
  label,
  valueText,
  tone = 'accent',
  indeterminate,
  className,
}: {
  value: number
  label: string
  valueText?: string
  tone?: 'accent' | 'success' | 'danger' | 'muted'
  indeterminate?: boolean
  className?: string
}) {
  const fills = {
    accent: 'bg-accent-gradient',
    success: 'bg-success',
    danger: 'bg-danger',
    muted: 'bg-line/25',
  }
  const pct = Math.max(0, Math.min(100, value))
  return (
    <div
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={indeterminate ? undefined : Math.round(pct)}
      aria-valuetext={valueText}
      className={cn('relative h-1.5 w-full overflow-hidden rounded-full bg-line/[0.08]', className)}
    >
      <motion.div
        className={cn('absolute inset-y-0 start-0 overflow-hidden rounded-full', fills[tone])}
        initial={false}
        animate={{ width: indeterminate ? '35%' : `${pct}%` }}
        transition={{ type: 'spring', stiffness: 120, damping: 24 }}
      >
        {tone === 'accent' && <Shimmer className="via-white/30" />}
      </motion.div>
    </div>
  )
}

export function Dialog({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
}: {
  open: boolean
  onOpenChange(v: boolean): void
  title: ReactNode
  description?: ReactNode
  children?: ReactNode
  footer?: ReactNode
}) {
  return (
    <RD.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <RD.Portal forceMount>
            <RD.Overlay asChild forceMount>
              <motion.div
                className="fixed inset-0 z-[80] bg-black/50 backdrop-blur-[2px]"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              />
            </RD.Overlay>
            <div className="pointer-events-none fixed inset-0 z-[81] flex items-center justify-center p-6">
            <RD.Content asChild forceMount dir={document.documentElement.dir as 'ltr' | 'rtl'}>
              <motion.div
                className="pointer-events-auto w-[min(480px,100%)] rounded-2xl border bg-elevated p-6 shadow-lift"
                initial={{ opacity: 0, scale: 0.96 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.97 }}
                transition={{ type: 'spring', stiffness: 420, damping: 32 }}
              >
                <RD.Title className="text-base font-semibold text-fg">{title}</RD.Title>
                {description ? (
                  <RD.Description className="mt-1.5 text-[13px] text-muted">{description}</RD.Description>
                ) : (
                  <RD.Description className="sr-only">{title}</RD.Description>
                )}
                {children && <div className="mt-4">{children}</div>}
                {footer && <div className="mt-6 flex justify-end gap-2">{footer}</div>}
              </motion.div>
            </RD.Content>
            </div>
          </RD.Portal>
        )}
      </AnimatePresence>
    </RD.Root>
  )
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  onConfirm,
  danger,
}: {
  open: boolean
  onOpenChange(v: boolean): void
  title: string
  description: string
  confirmLabel: string
  onConfirm(): void
  danger?: boolean
}) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      description={description}
      footer={
        <>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button
            variant={danger ? 'danger' : 'primary'}
            onClick={() => {
              onConfirm()
              onOpenChange(false)
            }}
          >
            {confirmLabel}
          </Button>
        </>
      }
    />
  )
}

export function Toaster() {
  const toasts = useUI((s) => s.toasts)
  const dismiss = useUI((s) => s.dismiss)
  const { t } = useTranslation()
  const icons = {
    success: <CheckCircle2 className="size-5 text-success" aria-hidden />,
    error: <AlertTriangle className="size-5 text-danger" aria-hidden />,
    info: <Info className="size-5 text-accent-ink" aria-hidden />,
  }
  return (
    <div className="pointer-events-none fixed bottom-5 end-5 z-[95] flex w-[min(380px,calc(100vw-40px))] flex-col gap-2" aria-live="polite" role="status">
      <AnimatePresence initial={false}>
        {toasts.map((x) => (
          <motion.div
            key={x.id}
            layout
            initial={{ opacity: 0, y: 16, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, scale: 0.96, transition: { duration: 0.15 } }}
            transition={{ type: 'spring', stiffness: 420, damping: 32 }}
            className="glass pointer-events-auto flex items-start gap-3 rounded-2xl border p-3.5 shadow-lift"
          >
            <div className="mt-0.5">{icons[x.kind]}</div>
            <div className="min-w-0 flex-1">
              <div className="text-[13px] font-semibold text-fg">{x.title}</div>
              {x.body && (
                <div className="mt-0.5 line-clamp-2 text-xs text-muted" dir="auto">
                  {x.body}
                </div>
              )}
              {x.action && (
                <button
                  type="button"
                  onClick={() => {
                    x.action!.run()
                    dismiss(x.id)
                  }}
                  className="mt-1.5 text-xs font-semibold text-accent-ink hover:underline"
                >
                  {x.action.label}
                </button>
              )}
            </div>
            <button type="button" aria-label={t('common.close')} onClick={() => dismiss(x.id)} className="rounded-md p-0.5 text-subtle hover:text-fg">
              <X className="size-4" />
            </button>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  )
}

/** Card section with a title, used in Settings and result panels. */
export function Section({ title, description, children, id, className }: { title?: ReactNode; description?: ReactNode; children: ReactNode; id?: string; className?: string }) {
  return (
    <section id={id} className={cn('card p-5', className)} aria-labelledby={id ? `${id}-title` : undefined}>
      {title && (
        <div className="mb-4">
          <h2 id={id ? `${id}-title` : undefined} className="text-[15px] font-semibold text-fg">
            {title}
          </h2>
          {description && <p className="mt-1 text-[13px] leading-relaxed text-muted">{description}</p>}
        </div>
      )}
      {children}
    </section>
  )
}

/** A settings row: label + hint on the start side, control on the end side. */
export function Row({ label, hint, children, htmlFor, stack }: { label: ReactNode; hint?: ReactNode; children: ReactNode; htmlFor?: string; stack?: boolean }) {
  return (
    <div className={cn('flex gap-4 border-t py-3.5 first:border-t-0 first:pt-0 last:pb-0', stack ? 'flex-col items-stretch' : 'items-center justify-between')}>
      <div className="min-w-0">
        <label htmlFor={htmlFor} className="block text-[13px] font-medium text-fg">
          {label}
        </label>
        {hint && <p className="mt-0.5 text-xs leading-relaxed text-muted">{hint}</p>}
      </div>
      <div className={cn('flex shrink-0 items-center gap-2', stack && 'w-full')}>{children}</div>
    </div>
  )
}
