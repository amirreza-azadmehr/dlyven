import { AnimatePresence, LayoutGroup, motion } from 'framer-motion'
import { ArrowLeft, FileDown, ShieldAlert } from 'lucide-react'
import { useState, type DragEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { cn } from '@/lib/cn'
import { findUrls } from '@/lib/platforms'
import { useAnalysis } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { PlatformIcon } from '@/components/Platform'
import { EngineBanner } from '@/components/EngineBanner'
import { ErrorPanel } from '@/components/ErrorPanel'
import { Button } from '@/components/ui/Button'
import { Kbd } from '@/components/ui/Controls'
import { BatchDialog } from './BatchDialog'
import { PreviewSkeleton } from './PreviewCard'
import { ResultView } from './ResultView'
import { UrlInput } from './UrlInput'

const FEATURED = ['youtube', 'instagram', 'tiktok', 'spotify', 'twitter', 'facebook', 'vimeo', 'soundcloud', 'reddit', 'twitch']

export function HomePage() {
  const { t } = useTranslation()
  const { status, info, error, analyze, reset, setUrl } = useAnalysis()
  const platforms = useSettings((s) => s.app?.platforms ?? [])
  const toast = useUI((s) => s.toast)
  const [batch, setBatch] = useState<string[]>([])
  const [dragging, setDragging] = useState(false)
  const hero = status === 'idle' || (status === 'error' && !info)
  const mac = navigator.platform.toLowerCase().includes('mac')

  const onDrop = async (e: DragEvent) => {
    e.preventDefault()
    setDragging(false)
    let text = e.dataTransfer.getData('text/uri-list') || e.dataTransfer.getData('text/plain')
    const files = [...e.dataTransfer.files].filter((f) => /\.(txt|list|csv|m3u8?)$/i.test(f.name) || f.type.startsWith('text/'))
    for (const f of files) text += '\n' + (await f.text())
    const urls = text ? await api.ExtractURLs(text).catch(() => findUrls(text)) : []
    if (urls.length === 0) return toast({ kind: 'info', title: t('home.dropNoLinks') })
    if (urls.length > 1) return setBatch(urls)
    setUrl(urls[0])
    void analyze(urls[0])
  }

  return (
    <div
      className="relative h-full"
      onDragOver={(e) => {
        e.preventDefault()
        e.dataTransfer.dropEffect = 'copy'
        if (!dragging) setDragging(true)
      }}
      onDragLeave={(e) => {
        if (e.currentTarget === e.target || !e.currentTarget.contains(e.relatedTarget as Node)) setDragging(false)
      }}
      onDrop={(e) => void onDrop(e)}
    >
      <LayoutGroup>
        <div className={cn('flex w-full flex-col', hero ? 'mx-auto min-h-full max-w-[880px] justify-center px-8 pb-16' : 'pb-6')}>
          {hero ? (
            <div className="flex flex-col items-center text-center">
              <motion.h1
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                className="text-[40px] font-bold leading-tight tracking-tight"
              >
                <span className="text-accent-gradient">{t('home.title')}</span>
              </motion.h1>
              <motion.p initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="mt-3 max-w-xl text-[15px] text-muted">
                {t('home.subtitle')}
              </motion.p>
              <div className="mt-9 w-full">
                <UrlInput onMany={setBatch} />
              </div>
              <motion.p initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.15 }} className="mt-4 flex items-center gap-2 text-xs text-subtle">
                <FileDown className="size-3.5" aria-hidden />
                {t('home.dropHint')}
              </motion.p>

              <div className="mt-6 w-full space-y-3 text-start">
                <EngineBanner />
                {error && <ErrorPanel error={error} onRetry={() => void analyze()} />}
              </div>

              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2 }} className="mt-10">
                <div className="mb-3 text-xs font-medium text-subtle">{t('home.worksWith')}</div>
                <ul className="flex max-w-xl flex-wrap justify-center gap-2">
                  {FEATURED.map((id) => platforms.find((p) => p.id === id))
                    .filter(Boolean)
                    .map((p, i) => (
                      <motion.li
                        key={p!.id}
                        initial={{ opacity: 0, y: 6 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.22 + i * 0.03 }}
                        className="flex h-8 items-center gap-1.5 rounded-full border bg-surface/60 px-3 text-xs text-muted"
                      >
                        <span style={{ color: p!.color }} className="flex">
                          <PlatformIcon icon={p!.icon} className="size-3.5" />
                        </span>
                        <bdi>{p!.name}</bdi>
                      </motion.li>
                    ))}
                  <li className="flex h-8 items-center px-2 text-xs text-subtle">{t('home.andMore')}</li>
                </ul>
              </motion.div>

              <p className="mt-10 flex items-center gap-1.5 text-xs text-subtle">
                {t('home.shortcutHint', { keys: '\u0000' })
                  .split('\u0000')
                  .flatMap((part, i) => (i === 0 ? [<span key={i}>{part}</span>] : [<Kbd key={`k${i}`}>{mac ? '⌘ V' : 'Ctrl V'}</Kbd>, <span key={i}>{part}</span>]))}
              </p>
              <p className="mt-3 flex items-center gap-1.5 text-2xs text-subtle">
                <ShieldAlert className="size-3" aria-hidden />
                {t('app.legal')}
              </p>
            </div>
          ) : (
            <div>
              {/* Full-width frosted toolbar so scrolled content blurs underneath without hard edges. */}
              <div className="sticky top-0 z-20 mb-4 border-b border-transparent bg-bg/85 backdrop-blur-xl [mask-image:linear-gradient(to_bottom,black_80%,transparent)]">
                <div className="mx-auto flex max-w-[880px] items-center gap-2 px-8 pb-4 pt-1">
                <Button variant="ghost" size="md" onClick={reset} icon={<ArrowLeft className="size-4 rtl:-scale-x-100" aria-hidden />} aria-label={t('home.newLink')}>
                  <span className="hidden md:inline">{t('home.newLink')}</span>
                </Button>
                <div className="flex-1">
                  <UrlInput compact onMany={setBatch} />
                </div>
                </div>
              </div>
              <div className="mx-auto max-w-[880px] px-8">
              <AnimatePresence mode="wait">
                {status === 'analyzing' && (
                  <motion.div key="sk" exit={{ opacity: 0 }}>
                    <PreviewSkeleton />
                  </motion.div>
                )}
                {status === 'ready' && info && <ResultView key={info.url} info={info} />}
                {status === 'error' && error && (
                  <motion.div key="err" className="pt-6">
                    <ErrorPanel error={error} onRetry={() => void analyze()} />
                  </motion.div>
                )}
              </AnimatePresence>
              </div>
            </div>
          )}
        </div>
      </LayoutGroup>

      <AnimatePresence>
        {dragging && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="pointer-events-none absolute inset-3 z-40 flex items-center justify-center rounded-3xl border-2 border-dashed border-accent/60 bg-accent/[0.06] backdrop-blur-sm"
          >
            <div className="flex flex-col items-center gap-3 text-accent-ink">
              <FileDown className="size-10" aria-hidden />
              <span className="text-[15px] font-semibold">{t('home.dropOverlay')}</span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <BatchDialog urls={batch} onClose={() => setBatch([])} />
    </div>
  )
}
