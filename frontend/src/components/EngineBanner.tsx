import { AnimatePresence, motion } from 'framer-motion'
import { Check, Cpu, Loader2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useFormat } from '@/lib/format'
import { useEngine } from '@/store/engine'
import { ProgressBar } from './ui/Feedback'
import { ErrorPanel } from './ErrorPanel'

const COMPONENTS = ['yt-dlp', 'ffmpeg', 'deno'] as const

/** First-run setup card: shows per-component download/verify progress. */
export function EngineBanner() {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { status, progress, error, install } = useEngine()
  const installing = Object.values(progress).some((p) => p.stage !== 'done' && p.stage !== 'error') || status?.busy
  const show = (status && !status.ready) || installing

  return (
    <AnimatePresence>
      {show && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          className="w-full overflow-hidden"
        >
          {error && !installing ? (
            <ErrorPanel error={error} onRetry={() => void install()} />
          ) : (
            <div className="card p-5" role="status" aria-live="polite">
              <div className="flex items-start gap-3.5">
                <div className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-accent/10">
                  <Cpu className="size-[18px] text-accent-ink" aria-hidden />
                </div>
                <div className="min-w-0 flex-1">
                  <h3 className="text-[14px] font-semibold">{t('engine.settingUp')}</h3>
                  <p className="mt-0.5 text-[13px] text-muted">{t('engine.settingUpBody')}</p>
                  <ul className="mt-4 space-y-3">
                    {COMPONENTS.map((c) => {
                      const p = progress[c]
                      const installed = status?.[c === 'yt-dlp' ? 'ytDlp' : c]?.installed
                      const stage = p?.stage ?? (installed ? 'done' : 'checking')
                      return (
                        <li key={c} className="text-[13px]">
                          <div className="mb-1.5 flex items-center justify-between gap-3">
                            <span className="flex items-center gap-2 font-medium" dir="ltr">
                              {stage === 'done' ? (
                                <Check className="size-3.5 text-success" aria-hidden />
                              ) : stage === 'error' ? (
                                <X className="size-3.5 text-danger" aria-hidden />
                              ) : (
                                <Loader2 className="size-3.5 animate-spin text-accent-ink" aria-hidden />
                              )}
                              {t(`settings.component.${c}`)}
                            </span>
                            <span className="tabular text-xs text-muted">
                              {t(`engine.stages.${stage}`)}
                              {stage === 'downloading' && p?.total ? ` · ${fmt.bytes(p.downloaded)} / ${fmt.bytes(p.total)}` : ''}
                            </span>
                          </div>
                          <ProgressBar
                            value={stage === 'done' ? 100 : p?.percent ?? 0}
                            indeterminate={stage === 'checking' || (stage === 'downloading' && !p?.total)}
                            tone={stage === 'done' ? 'success' : stage === 'error' ? 'danger' : 'accent'}
                            label={t(`settings.component.${c}`)}
                            valueText={t(`engine.stages.${stage}`)}
                          />
                        </li>
                      )
                    })}
                  </ul>
                </div>
              </div>
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  )
}
