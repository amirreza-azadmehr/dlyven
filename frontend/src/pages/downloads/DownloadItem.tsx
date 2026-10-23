import { motion } from 'framer-motion'
import { Check, Copy, ExternalLink, FolderOpen, Music, Pause, Play, RotateCcw, ShieldCheck, Trash2, X } from 'lucide-react'
import { memo } from 'react'
import { useTranslation } from 'react-i18next'
import { api, rt, toAppError } from '@/lib/api'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { Job } from '@/lib/types'
import { isActive } from '@/store/downloads'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { PlatformBadge } from '@/components/Platform'
import { IconButton } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Controls'
import { ProgressBar } from '@/components/ui/Feedback'
import { errorHint, errorTitle } from '@/components/ErrorPanel'
import { BidiList } from '@/components/ui/BidiList'

const tones: Record<Job['status'], 'neutral' | 'accent' | 'success' | 'warning' | 'danger'> = {
  queued: 'neutral',
  downloading: 'accent',
  merging: 'accent',
  converting: 'accent',
  paused: 'warning',
  retrying: 'warning',
  done: 'success',
  failed: 'danger',
  canceled: 'neutral',
}

export const DownloadItem = memo(function DownloadItem({ job }: { job: Job }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const toast = useUI((s) => s.toast)
  const registry = useSettings((s) => s.app?.platforms ?? [])
  const r = job.request
  const platform = registry.find((p) => p.id === r.platform)
  const active = isActive(job)
  const square = !!r.spotify || r.kind === 'audio'

  const act = (fn: () => Promise<unknown>) => () =>
    void fn().catch((e) => {
      const err = toAppError(e)
      toast({ kind: 'error', title: errorTitle(t, err.code), body: errorHint(t, err.code) })
    })

  const statusLine = (): string[] => {
    switch (job.status) {
      case 'downloading': {
        const parts: string[] = []
        if (job.total > 0 && job.downloaded >= 0) parts.push(t('downloads.sizeOf', { done: fmt.bytes(job.downloaded), total: fmt.bytes(job.total) }))
        else if (job.downloaded > 0) parts.push(fmt.bytes(job.downloaded))
        if (job.speed > 0) parts.push(t('downloads.perSecond', { speed: fmt.speed(job.speed) }))
        if (job.eta >= 0) parts.push(t('downloads.eta', { time: fmt.duration(job.eta) }))
        if (job.parts > 1) parts.push(t('downloads.part', { part: fmt.number(job.part), parts: fmt.number(job.parts) }))
        return parts
      }
      case 'done':
        return [fmt.bytes(job.total), job.finishedAt ? fmt.dateTime(job.finishedAt) : ''].filter(Boolean)
      case 'failed':
        return [errorTitle(t, job.errorCode || 'ERR_UNKNOWN')]
      case 'retrying':
        return [errorTitle(t, job.errorCode || 'ERR_UNKNOWN'), t('downloads.attempt', { n: fmt.number(job.attempts + 1) })]
      case 'paused':
        return job.total > 0 ? [t('downloads.sizeOf', { done: fmt.bytes(job.downloaded), total: fmt.bytes(job.total) })] : []
      default:
        return []
    }
  }

  const shownProgress = !['done', 'failed', 'canceled'].includes(job.status) ? Math.min(job.progress, 99) : job.progress
  const progressText = `${fmt.percent(shownProgress)} · ${t(`downloads.status.${job.status}`)}`

  return (
    <motion.li
      layout
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.98, transition: { duration: 0.15 } }}
      transition={{ type: 'spring', stiffness: 380, damping: 34 }}
      className="card group flex items-center gap-4 p-3"
    >
      <div className={cn('relative shrink-0 overflow-hidden rounded-xl bg-line/[0.06]', square ? 'size-14' : 'h-14 w-24')}>
        {r.thumbnail ? (
          <img src={r.thumbnail} alt="" loading="lazy" referrerPolicy="no-referrer" className="size-full object-cover" />
        ) : (
          <div className="flex size-full items-center justify-center text-subtle">
            <Music className="size-5" />
          </div>
        )}
        {job.status === 'done' && (
          <motion.span
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ type: 'spring', stiffness: 500, damping: 22 }}
            className="absolute inset-0 flex items-center justify-center bg-black/35"
          >
            <span className="flex size-7 items-center justify-center rounded-full bg-success text-white shadow-lg">
              <Check className="size-4" strokeWidth={3} />
            </span>
          </motion.span>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <h3 className="min-w-0 truncate text-[13px] font-semibold text-fg" dir="auto" title={r.title}>
            <bdi>{r.title || r.url}</bdi>
          </h3>
        </div>
        <div className="mt-1 flex flex-wrap items-center gap-1.5">
          <Badge tone={tones[job.status]}>{t(`downloads.status.${job.status}`)}</Badge>
          {platform && <PlatformBadge platform={{ ...platform, known: true }} size="sm" className="h-5" />}
          {r.qualityLabel && (
            <Badge>
              <bdi dir="ltr">{r.qualityLabel}</bdi>
            </Badge>
          )}
          {r.spotify && (
            <Badge tone="success" title={t('result.spotify.title')}>
              <ShieldCheck className="size-3" /> {t('downloads.matched')}
            </Badge>
          )}
          <span className="tabular ms-1 truncate text-xs text-muted" title={job.status === 'failed' ? errorHint(t, job.errorCode) : undefined}>
            <BidiList items={statusLine()} />
          </span>
        </div>
        {!['done', 'failed', 'canceled'].includes(job.status) && (
          <div className="mt-2.5 flex items-center gap-3">
            <ProgressBar
              value={shownProgress}
              indeterminate={job.status === 'queued' || job.status === 'merging' || job.status === 'converting' || (job.status === 'downloading' && job.progress <= 0 && job.total <= 0)}
              tone={job.status === 'paused' || job.status === 'queued' ? 'muted' : 'accent'}
              label={t('downloads.progressLabel', { title: r.title })}
              valueText={progressText}
            />
            <span className="tabular w-10 shrink-0 text-end text-xs font-medium text-muted">{fmt.percent(shownProgress)}</span>
          </div>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-0.5 opacity-80 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
        {(active || job.status === 'queued') && (
          <IconButton label={t('downloads.actions.pause')} onClick={act(() => api.PauseDownload(job.id))}>
            <Pause className="size-4" />
          </IconButton>
        )}
        {job.status === 'paused' && (
          <IconButton label={t('downloads.actions.resume')} onClick={act(() => api.ResumeDownload(job.id))}>
            <Play className="size-4" />
          </IconButton>
        )}
        {(job.status === 'failed' || job.status === 'canceled') && (
          <IconButton label={t('downloads.actions.retry')} onClick={act(() => api.RetryDownload(job.id))}>
            <RotateCcw className="size-4" />
          </IconButton>
        )}
        {job.status === 'done' && job.filePath && (
          <>
            <IconButton label={t('downloads.actions.openFile')} onClick={act(() => api.OpenFile(job.filePath))}>
              <ExternalLink className="size-4 rtl:-scale-x-100" />
            </IconButton>
            <IconButton label={t('downloads.actions.showInFolder')} onClick={act(() => api.ShowInFolder(job.filePath))}>
              <FolderOpen className="size-4" />
            </IconButton>
          </>
        )}
        {r.url && (
          <IconButton
            label={t('downloads.actions.copyLink')}
            onClick={act(async () => {
              await rt.ClipboardSetText(r.url)
              toast({ kind: 'success', title: t('common.copied') })
            })}
          >
            <Copy className="size-4" />
          </IconButton>
        )}
        {!['done', 'failed', 'canceled'].includes(job.status) ? (
          <IconButton label={t('downloads.actions.cancel')} tone="danger" onClick={act(() => api.CancelDownload(job.id))}>
            <X className="size-4" />
          </IconButton>
        ) : (
          <IconButton label={t('downloads.actions.remove')} tone="danger" onClick={act(() => api.RemoveDownload(job.id))}>
            <Trash2 className="size-4" />
          </IconButton>
        )}
      </div>
    </motion.li>
  )
})
