import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { MediaInfo } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { Checkbox } from '@/components/ui/Controls'

/** Checklist for playlists / albums / carousels. */
export function PlaylistPanel({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const selected = new Set(sel.entries)
  const all = info.entries.length
  const toggle = (i: number) => {
    const next = new Set(selected)
    if (next.has(i)) next.delete(i)
    else next.add(i)
    select({ entries: [...next].sort((a, b) => a - b) })
  }

  return (
    <section className="card overflow-hidden" aria-label={t('result.playlist.items')}>
      <div className="flex items-center justify-between gap-3 border-b px-4 py-2.5">
        <label className="flex items-center gap-3 text-[13px] font-medium">
          <Checkbox
            label={t('common.selectAll')}
            checked={selected.size === all}
            indeterminate={selected.size > 0 && selected.size < all}
            onChange={(v) => select({ entries: v ? info.entries.map((e) => e.index) : [] })}
          />
          {t('result.playlist.selected', { n: fmt.number(selected.size), total: fmt.number(all) })}
        </label>
        <div className="flex gap-1">
          <button type="button" className="rounded-lg px-2 py-1 text-xs font-medium text-accent-ink hover:bg-accent/10" onClick={() => select({ entries: info.entries.map((e) => e.index) })}>
            {t('common.selectAll')}
          </button>
          <button type="button" className="rounded-lg px-2 py-1 text-xs font-medium text-muted hover:bg-line/[0.06]" onClick={() => select({ entries: [] })}>
            {t('common.selectNone')}
          </button>
        </div>
      </div>
      <ul className="max-h-[300px] overflow-y-auto p-1.5">
        {info.entries.map((e, i) => (
          <motion.li
            key={e.index}
            initial={{ opacity: 0, x: 8 * (document.documentElement.dir === 'rtl' ? -1 : 1) }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: Math.min(i, 12) * 0.025 }}
          >
            <label
              className={cn(
                'flex cursor-pointer items-center gap-3 rounded-xl px-2.5 py-2 transition-colors hover:bg-line/[0.05]',
                !selected.has(e.index) && 'opacity-60',
              )}
            >
              <Checkbox label={e.title} checked={selected.has(e.index)} onChange={() => toggle(e.index)} />
              <span className="tabular w-6 shrink-0 text-end text-xs text-subtle">{fmt.number(e.index)}</span>
              {e.thumbnail ? (
                <img src={e.thumbnail} alt="" loading="lazy" referrerPolicy="no-referrer" className={cn('shrink-0 rounded-md object-cover', info.matched ? 'size-9' : 'h-9 w-16')} />
              ) : (
                <span className={cn('shrink-0 rounded-md bg-line/[0.07]', info.matched ? 'size-9' : 'h-9 w-16')} />
              )}
              <span className="min-w-0 flex-1">
                <span className="block truncate text-[13px] text-fg" dir="auto">
                  {e.title || t('result.playlist.untitled')}
                </span>
                {e.uploader && (
                  <span className="block truncate text-xs text-muted" dir="auto">
                    {e.uploader}
                  </span>
                )}
              </span>
              {e.duration > 0 && (
                <span dir="ltr" className="tabular shrink-0 text-xs text-subtle">
                  {fmt.duration(e.duration)}
                </span>
              )}
            </label>
          </motion.li>
        ))}
      </ul>
    </section>
  )
}
