import { motion } from 'framer-motion'
import { Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { MediaInfo } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { Segmented } from '@/components/ui/Controls'

export function ThumbnailTab({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const thumbs = info.thumbnails.length ? info.thumbnails : info.thumbnail ? [{ url: info.thumbnail, width: 0, height: 0, id: 'main' }] : []
  const square = info.matched || info.platform.audioOnly

  return (
    <div>
      <div role="radiogroup" aria-label={t('result.thumbnail.size')} className="grid max-h-[340px] grid-cols-3 gap-3 overflow-y-auto p-4">
        {thumbs.map((th, i) => {
          const active = (sel.thumbUrl || thumbs[0]?.url) === th.url
          const dims = th.width ? t('result.thumbnail.dimensions', { w: fmt.number(th.width, { useGrouping: false }), h: fmt.number(th.height, { useGrouping: false }) }) : t('common.unknown')
          return (
            <motion.button
              key={th.url + i}
              type="button"
              role="radio"
              aria-checked={active}
              aria-label={dims}
              initial={{ opacity: 0, scale: 0.97 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: Math.min(i, 9) * 0.03 }}
              onClick={() => select({ thumbUrl: th.url })}
              className={cn('group overflow-hidden rounded-xl border-2 text-start transition-colors', active ? 'border-accent' : 'border-transparent hover:border-line/20')}
            >
              <div className={cn('relative bg-line/[0.06]', square ? 'aspect-square' : 'aspect-video')}>
                <img src={th.url} alt="" loading="lazy" referrerPolicy="no-referrer" className="size-full object-cover" />
                {active && (
                  <span className="absolute end-2 top-2 flex size-6 items-center justify-center rounded-full bg-accent text-accent-fg shadow">
                    <Check className="size-3.5" strokeWidth={3} />
                  </span>
                )}
              </div>
              <div className="tabular px-2 py-1.5 text-xs text-muted" dir="ltr">
                {dims}
              </div>
            </motion.button>
          )
        })}
      </div>
      <div className="flex items-center justify-between gap-3 border-t px-4 py-3">
        <span className="text-[13px] font-medium">{t('result.thumbnail.format')}</span>
        <Segmented
          size="sm"
          label={t('result.thumbnail.format')}
          value={sel.thumbFormat}
          onChange={(thumbFormat) => select({ thumbFormat })}
          options={[
            { value: 'original', label: t('result.thumbnail.original') },
            { value: 'jpg', label: 'JPG' },
            { value: 'png', label: 'PNG' },
          ]}
        />
      </div>
    </div>
  )
}
