import { AnimatePresence, motion } from 'framer-motion'
import { ArrowRight, ClipboardPaste, Link2, Loader2, X } from 'lucide-react'
import { useMemo, useRef, type KeyboardEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { rt } from '@/lib/api'
import { cn } from '@/lib/cn'
import { detectPlatform, findUrls } from '@/lib/platforms'
import { useAnalysis } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { PlatformBadge } from '@/components/Platform'
import { Button, IconButton } from '@/components/ui/Button'
import { Kbd } from '@/components/ui/Controls'

/**
 * Command-palette style link input. It carries layoutId="url-bar" so it glides from the
 * centered hero position to the top of the result view (shared-element transition).
 */
export function UrlInput({ compact, onMany }: { compact?: boolean; onMany(urls: string[]): void }) {
  const { t } = useTranslation()
  const { url, status, setUrl, analyze, cancel, reset } = useAnalysis()
  const registry = useSettings((s) => s.app?.platforms ?? [])
  const inputRef = useRef<HTMLInputElement>(null)
  const platform = useMemo(() => detectPlatform(url, registry), [url, registry])
  const analyzing = status === 'analyzing'
  const invalid = url.trim().length > 6 && !platform.valid

  const submit = (value = url) => {
    const urls = findUrls(value)
    if (urls.length > 1) return onMany(urls)
    if (detectPlatform(value, registry).valid) void analyze(value)
  }

  const paste = async () => {
    const text = (await rt.ClipboardGetText().catch(() => '')) ?? ''
    if (!text.trim()) return inputRef.current?.focus()
    const urls = findUrls(text)
    if (urls.length > 1) return onMany(urls)
    setUrl(urls[0] ?? text.trim())
    submit(urls[0] ?? text.trim())
  }

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      submit()
    }
    if (e.key === 'Escape') {
      e.stopPropagation()
      if (analyzing) cancel()
      else if (url) setUrl('')
    }
  }

  return (
    <motion.div
      layoutId="url-bar"
      transition={{ type: 'spring', stiffness: 260, damping: 30 }}
      className={cn(
        'glass group relative flex w-full items-center gap-2 border shadow-lift transition-[box-shadow,border-color]',
        'focus-within:border-accent/40 focus-within:shadow-glow',
        compact ? 'h-12 rounded-2xl ps-3.5 pe-1.5' : 'h-16 rounded-[20px] ps-5 pe-2',
        invalid && 'border-danger/40',
      )}
    >
      <Link2 className={cn('shrink-0 text-subtle', compact ? 'size-4' : 'size-5')} aria-hidden />
      <AnimatePresence mode="popLayout" initial={false}>
        {platform.valid && (
          <motion.div key={platform.id + platform.name} initial={{ opacity: 0, scale: 0.85 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.85 }}>
            <PlatformBadge platform={platform} size={compact ? 'sm' : 'md'} />
          </motion.div>
        )}
      </AnimatePresence>
      <input
        ref={inputRef}
        type="url"
        inputMode="url"
        // URLs are always LTR; while empty the placeholder follows the UI direction.
        dir={url ? 'ltr' : undefined}
        autoFocus={!compact}
        spellCheck={false}
        autoComplete="off"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        onKeyDown={onKeyDown}
        onPaste={(e) => {
          const text = e.clipboardData.getData('text')
          const urls = findUrls(text)
          if (urls.length > 1) {
            e.preventDefault()
            onMany(urls)
          } else if (!url.trim() && urls.length === 1) {
            e.preventDefault()
            setUrl(urls[0])
            submit(urls[0])
          }
        }}
        placeholder={t('home.placeholder')}
        aria-label={t('home.inputLabel')}
        aria-invalid={invalid || undefined}
        aria-describedby={invalid ? 'url-error' : undefined}
        className={cn(
          'no-drag h-full min-w-0 flex-1 bg-transparent text-start text-fg outline-none placeholder:text-subtle',
          compact ? 'text-[13px]' : 'text-[15px]',
        )}
      />
      {url && !analyzing && (
        <IconButton label={t('home.clearInput')} size="sm" onClick={() => (compact ? reset() : setUrl(''))}>
          <X className="size-4" />
        </IconButton>
      )}
      {!url && !compact && (
        <Button variant="ghost" size="md" icon={<ClipboardPaste className="size-4" aria-hidden />} onClick={() => void paste()}>
          {t('home.paste')}
          <span className="ms-1 hidden sm:inline-flex">
            <Kbd>{navigator.platform.toLowerCase().includes('mac') ? '⌘V' : 'Ctrl V'}</Kbd>
          </span>
        </Button>
      )}
      {analyzing ? (
        <Button variant="secondary" size={compact ? 'md' : 'lg'} onClick={cancel} icon={<Loader2 className="size-4 animate-spin" aria-hidden />} aria-label={t('home.cancelAnalyze')}>
          {t('common.cancel')}
        </Button>
      ) : (
        <Button
          variant="primary"
          size={compact ? 'md' : 'lg'}
          disabled={!platform.valid}
          onClick={() => submit()}
          iconEnd={<ArrowRight className="size-4 rtl:-scale-x-100" aria-hidden />}
        >
          {t('home.analyze')}
        </Button>
      )}
      <AnimatePresence>
        {invalid && (
          <motion.p
            id="url-error"
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            className="absolute -bottom-6 start-5 text-xs text-danger"
          >
            {t('home.invalidUrl')}
          </motion.p>
        )}
      </AnimatePresence>
    </motion.div>
  )
}
