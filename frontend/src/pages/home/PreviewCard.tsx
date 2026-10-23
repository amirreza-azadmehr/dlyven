import { motion } from 'framer-motion'
import { CalendarDays, Clock3, ExternalLink, Eye, ListVideo, Music, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { rt } from '@/lib/api'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { MediaInfo } from '@/lib/types'
import { PlatformBadge } from '@/components/Platform'
import { IconButton } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Controls'

export function PreviewCard({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const square = info.matched || info.platform.audioOnly
  const count = info.entries.length

  return (
    <motion.article
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      className="card flex gap-5 p-4"
    >
      <div className={cn('relative shrink-0 overflow-hidden rounded-xl bg-line/[0.06]', square ? 'size-[132px]' : 'aspect-video w-[236px]')}>
        {info.thumbnail ? (
          <img src={info.thumbnail} alt="" className="size-full object-cover" loading="lazy" referrerPolicy="no-referrer" />
        ) : (
          <div className="flex size-full items-center justify-center text-subtle">
            {square ? <Music className="size-8" /> : <ListVideo className="size-8" />}
          </div>
        )}
        {info.duration > 0 && !info.isPlaylist && (
          <span dir="ltr" className="tabular absolute bottom-1.5 end-1.5 rounded-md bg-black/70 px-1.5 py-0.5 text-2xs font-semibold text-white">
            {fmt.duration(info.duration)}
          </span>
        )}
      </div>
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex items-start justify-between gap-3">
          <PlatformBadge platform={info.platform} size="sm" />
          <IconButton label={t('result.openSource')} size="sm" onClick={() => rt.BrowserOpenURL(info.url)}>
            <ExternalLink className="size-3.5 rtl:-scale-x-100" />
          </IconButton>
        </div>
        <h2 className="mt-2 line-clamp-2 text-[17px] font-semibold leading-snug text-fg" dir="auto" data-selectable>
          <bdi>{info.title}</bdi>
        </h2>
        {info.uploader && (
          <p className="mt-1 truncate text-[13px] text-muted" dir="auto">
            {t('result.by', { name: '⁨' + info.uploader + '⁩' })}
          </p>
        )}
        <div className="mt-auto flex flex-wrap items-center gap-x-4 gap-y-1 pt-3 text-xs text-muted">
          {info.isPlaylist && (
            <Meta icon={<ListVideo className="size-3.5" />}>{t('result.playlist', { count, n: fmt.number(count) })}</Meta>
          )}
          {info.duration > 0 && (
            <Meta icon={<Clock3 className="size-3.5" />} label={t('result.duration')}>
              <span dir="ltr">{fmt.duration(info.duration)}</span>
            </Meta>
          )}
          {info.uploadDate && (
            <Meta icon={<CalendarDays className="size-3.5" />} label={t('result.uploaded')}>
              {fmt.date(info.uploadDate)}
            </Meta>
          )}
          {info.viewCount > 0 && (
            <Meta icon={<Eye className="size-3.5" />}>{t('result.views', { count: info.viewCount, n: fmt.compact(info.viewCount) })}</Meta>
          )}
          {info.matched && (
            <Meta icon={<ShieldCheck className="size-3.5 text-success" />}>{t('home.matchedSource')}</Meta>
          )}
        </div>
      </div>
    </motion.article>
  )
}

function Meta({ icon, children, label }: { icon: React.ReactNode; children: React.ReactNode; label?: string }) {
  return (
    <span className="inline-flex items-center gap-1.5" title={label}>
      <span className="text-subtle" aria-hidden>
        {icon}
      </span>
      {label && <span className="sr-only">{label}:</span>}
      {children}
    </span>
  )
}

export function PreviewSkeleton() {
  return (
    <div className="space-y-4" aria-busy="true">
      <div className="card flex gap-5 p-4">
        <Skeleton className="aspect-video w-[236px] shrink-0 rounded-xl" />
        <div className="flex flex-1 flex-col gap-2.5 pt-1">
          <Skeleton className="h-6 w-24 rounded-full" />
          <Skeleton className="h-5 w-4/5" />
          <Skeleton className="h-5 w-2/5" />
          <Skeleton className="mt-auto h-4 w-1/2" />
        </div>
      </div>
      <div className="flex gap-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-9 w-24 rounded-xl" />
        ))}
      </div>
      <div className="card space-y-3 p-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-11 w-full rounded-xl" />
        ))}
      </div>
    </div>
  )
}
