import { motion } from 'framer-motion'
import { Volume2, VolumeX } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import { findFormat } from '@/lib/requests'
import type { Container, Format, MediaInfo } from '@/lib/types'
import { useAnalysis, type PresetId } from '@/store/analysis'
import { Badge, Segmented } from '@/components/ui/Controls'
import { Tooltip } from '@/components/ui/Tooltip'

const WEBM_CODECS = ['VP9', 'VP8', 'AV1']

/** Keeps the presets row in sync when a playlist quality cap is chosen. */
function presetForCap(v: string): PresetId | null {
  if (v === '0') return 'best'
  return (['2160', '1080', '720', '480'] as const).find((p) => p === v) ?? null
}

function resolutionTag(h: number): string | null {
  if (h >= 4320) return '8K'
  if (h >= 2160) return '4K'
  if (h >= 1440) return '2K'
  if (h >= 720) return 'HD'
  return null
}

export function VideoTab({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const current = findFormat(info, sel.formatId)
  const webmOk = info.isPlaylist || !current || WEBM_CODECS.includes(current.vcodec)
  const subsCount = new Set(sel.subs.map((s) => s.split('|')[0])).size

  const container = (
    <div className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
      <div className="min-w-0">
        <div className="text-[13px] font-medium">{t('result.video.container')}</div>
        <p className="text-xs text-muted">{t('result.video.containerHint')}</p>
      </div>
      <Segmented<Container>
        label={t('result.video.container')}
        value={sel.container === 'webm' && !webmOk ? 'mp4' : sel.container}
        onChange={(container) => select({ container })}
        options={[
          { value: 'mp4', label: 'MP4' },
          { value: 'mkv', label: 'MKV' },
          { value: 'webm', label: 'WebM', disabled: !webmOk },
        ]}
      />
    </div>
  )

  if (info.isPlaylist) {
    const caps = [0, 2160, 1440, 1080, 720, 480, 360]
    return (
      <div>
        <div className="p-4">
          <div className="mb-1 text-[13px] font-medium">{t('result.video.qualityCap')}</div>
          <p className="mb-3 text-xs text-muted">{t('result.video.playlistHint')}</p>
          <Segmented<string>
            label={t('result.video.qualityCap')}
            value={String(sel.maxHeight)}
            onChange={(v) => select({ maxHeight: Number(v), formatId: null, preset: presetForCap(v) })}
            options={caps.map((c) => ({ value: String(c), label: c === 0 ? t('result.presets.best') : c === 2160 ? '4K' : `${fmt.number(c, { useGrouping: false })}p` }))}
          />
        </div>
        {container}
      </div>
    )
  }

  if (info.videoGroups.length === 0) {
    return <p className="p-6 text-center text-[13px] text-muted">{t('result.video.noFormats')}</p>
  }

  return (
    <div>
      <div role="radiogroup" aria-label={t('result.tabs.video')} className="max-h-[420px] overflow-y-auto p-2">
        {info.videoGroups.map((g, gi) => (
          <motion.div key={g.height} initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: Math.min(gi, 8) * 0.03 }} className="mb-1.5">
            <div className="flex items-center gap-2 px-2.5 pb-1 pt-2">
              <span className="text-[13px] font-semibold tabular">{g.height ? `${fmt.number(g.height, { useGrouping: false })}p` : t('result.video.unknownResolution')}</span>
              {resolutionTag(g.height) && <Badge tone="accent">{resolutionTag(g.height)}</Badge>}
              <span className="text-xs text-subtle">· {t('result.video.groupStreams', { count: g.formats.length, n: fmt.number(g.formats.length) })}</span>
            </div>
            {g.formats.map((f) => (
              <FormatRow key={f.id} f={f} info={info} selected={sel.formatId === f.id} onSelect={() => select({ formatId: f.id, preset: null, tab: 'video', container: !WEBM_CODECS.includes(f.vcodec) && sel.container === 'webm' ? 'mp4' : sel.container })} />
            ))}
          </motion.div>
        ))}
      </div>
      {subsCount > 0 && sel.subMode === 'embed' && (
        <p className="border-t px-4 py-2 text-xs text-muted">{t('result.video.embedSubsNote', { count: subsCount, n: fmt.number(subsCount) })}</p>
      )}
      {container}
    </div>
  )
}

function FormatRow({ f, info, selected, onSelect }: { f: Format; info: MediaInfo; selected: boolean; onSelect(): void }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const h = info.highlights
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      onClick={onSelect}
      className={cn(
        'relative grid w-full grid-cols-[18px_minmax(88px,1.1fr)_minmax(110px,1.4fr)_64px_80px_84px] items-center gap-3 rounded-xl px-2.5 py-2 text-start text-[13px] transition-colors',
        selected ? 'text-fg' : 'text-muted hover:bg-line/[0.04] hover:text-fg',
      )}
    >
      {selected && (
        <motion.span layoutId="format-selected" className="absolute inset-0 rounded-xl border border-accent/40 bg-accent/[0.08]" transition={{ type: 'spring', stiffness: 500, damping: 40 }} />
      )}
      <span className={cn('relative flex size-[18px] items-center justify-center rounded-full border-2', selected ? 'border-accent' : 'border-line/25')}>
        {selected && <motion.span layoutId="format-dot" className="size-2 rounded-full bg-accent" />}
      </span>
      <span className="relative flex items-center gap-1.5 font-medium text-fg">
        <span className="tabular" dir="ltr">{fmt.digits(f.label || f.id)}</span>
        {f.hdr && <Badge tone="warning">{t('result.badges.hdr')}</Badge>}
      </span>
      <span className="relative flex min-w-0 flex-wrap items-center gap-1.5">
        <span dir="ltr" className="truncate">
          {f.vcodec || '—'}
          {f.hasAudio && f.acodec ? ` + ${f.acodec}` : ''}
        </span>
        {f.id === h.best && <Badge tone="accent">{t('result.badges.best')}</Badge>}
        {f.id === h.recommended && <Badge tone="success">{t('result.badges.recommended')}</Badge>}
        {f.id === h.smallest && f.id !== h.best && <Badge>{t('result.badges.smallest')}</Badge>}
      </span>
      <span className="relative flex items-center gap-1">
        <span className="uppercase">{f.ext}</span>
        <Tooltip content={f.hasAudio ? t('result.columns.audio') : t('result.badges.mergedHint')}>
          <span className="text-subtle" aria-label={f.hasAudio ? t('result.columns.audio') : t('result.badges.mergedHint')}>
            {f.hasAudio ? <Volume2 className="size-3.5" /> : <VolumeX className="size-3.5" />}
          </span>
        </Tooltip>
      </span>
      <span className="relative tabular text-end" title={t('result.columns.bitrate')}>
        {f.tbr ? fmt.bitrate(f.tbr) : '—'}
      </span>
      <span className="relative tabular text-end font-medium text-fg" title={f.sizeApprox ? t('result.badges.approx') : t('result.columns.size')}>
        {fmt.bytes(f.filesize, f.sizeApprox)}
      </span>
    </button>
  )
}
