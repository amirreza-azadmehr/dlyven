import { AlertTriangle, ChevronDown } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import type { AppError } from '@/lib/types'
import { useUI, type SettingsSection } from '@/store/ui'
import { useEngine } from '@/store/engine'
import { Button } from './ui/Button'

/** Maps an error code to the most helpful next step. */
function fixFor(code: string): { label: string; section?: SettingsSection; engine?: 'install' | 'update' } | null {
  switch (code) {
    case 'ERR_LOGIN_REQUIRED':
    case 'ERR_AGE_RESTRICTED':
    case 'ERR_PRIVATE':
    case 'ERR_COOKIES':
      return { label: 'errors.actions.openCookies', section: 'cookies' }
    case 'ERR_NETWORK':
    case 'ERR_TIMEOUT':
    case 'ERR_PROXY':
    case 'ERR_GEO_BLOCKED':
    case 'ERR_RATE_LIMITED':
      return { label: 'errors.actions.openNetwork', section: 'network' }
    case 'ERR_ENGINE_MISSING':
    case 'ERR_FFMPEG_MISSING':
    case 'ERR_ENGINE_INSTALL':
      return { label: 'errors.actions.installEngine', engine: 'install' }
    case 'ERR_UNKNOWN':
    case 'ERR_UNSUPPORTED_URL':
    case 'ERR_FORMAT_UNAVAILABLE':
      return { label: 'errors.actions.updateEngine', engine: 'update' }
  }
  return null
}

export function errorTitle(t: (k: string) => string, code: string) {
  const key = `errors.${code}.title`
  const v = t(key)
  return v === key ? t('errors.ERR_UNKNOWN.title') : v
}

export function errorHint(t: (k: string) => string, code: string) {
  const key = `errors.${code}.hint`
  const v = t(key)
  return v === key ? t('errors.ERR_UNKNOWN.hint') : v
}

export function ErrorPanel({ error, onRetry }: { error: AppError; onRetry?: () => void }) {
  const { t } = useTranslation()
  const navigate = useUI((s) => s.navigate)
  const engine = useEngine()
  const [open, setOpen] = useState(false)
  const fix = fixFor(error.code)

  return (
    <motion.div
      role="alert"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="card w-full border-danger/25 bg-danger/[0.04] p-5"
    >
      <div className="flex items-start gap-3.5">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-danger/10">
          <AlertTriangle className="size-[18px] text-danger" aria-hidden />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="text-[14px] font-semibold text-fg">{errorTitle(t, error.code)}</h3>
          <p className="mt-1 text-[13px] leading-relaxed text-muted">{errorHint(t, error.code)}</p>
          <div className="mt-4 flex flex-wrap gap-2">
            {fix && (
              <Button
                variant="primary"
                size="sm"
                loading={!!fix.engine && engine.working}
                onClick={() => {
                  if (fix.section) navigate('settings', fix.section)
                  else if (fix.engine === 'install') void engine.install().then(() => onRetry?.())
                  else if (fix.engine === 'update') void engine.update().then((ok) => ok && onRetry?.())
                }}
              >
                {t(fix.label)}
              </Button>
            )}
            {onRetry && (
              <Button size="sm" onClick={onRetry}>
                {t('errors.actions.retry')}
              </Button>
            )}
            {error.detail && (
              <button
                type="button"
                onClick={() => setOpen(!open)}
                aria-expanded={open}
                className="inline-flex items-center gap-1 rounded-lg px-2 text-xs font-medium text-subtle hover:text-fg"
              >
                {t('errors.details')}
                <ChevronDown className={`size-3.5 transition-transform ${open ? 'rotate-180' : ''}`} aria-hidden />
              </button>
            )}
          </div>
          <AnimatePresence>
            {open && error.detail && (
              <motion.pre
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                dir="ltr"
                data-selectable
                className="mt-3 overflow-hidden whitespace-pre-wrap break-words rounded-xl bg-line/[0.05] p-3 text-start font-mono text-2xs text-muted"
              >
                {error.detail}
              </motion.pre>
            )}
          </AnimatePresence>
        </div>
      </div>
    </motion.div>
  )
}
