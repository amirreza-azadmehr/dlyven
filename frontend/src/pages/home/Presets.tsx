import { motion } from 'framer-motion'
import { Info, Music, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import type { MediaInfo } from '@/lib/types'
import { useAnalysis, type PresetId } from '@/store/analysis'
import { Tooltip } from '@/components/ui/Tooltip'

const VIDEO: PresetId[] = ['best', '2160', '1080', '720', '480']
const AUDIO: PresetId[] = ['mp3', 'm4a']

export function PresetsRow({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const { sel, applyPreset } = useAnalysis()
  const videoAllowed = !info.matched && !info.platform.audioOnly && (info.isPlaylist || info.videoGroups.length > 0)
  const available = (p: PresetId) => {
    if (AUDIO.includes(p)) return true
    if (!videoAllowed) return false
    if (p === 'best' || info.isPlaylist) return true
    return info.maxHeight >= Number(p)
  }
  const presets = videoAllowed ? [...VIDEO, ...AUDIO] : AUDIO

  return (
    <div role="group" aria-label={t('result.presets.label')} className="flex flex-wrap items-center gap-2">
      <span className="me-1 text-xs font-medium text-subtle">{t('result.presets.label')}</span>
      {presets.map((p) => {
        const active = sel.preset === p
        const ok = available(p)
        const btn = (
          <button
            key={p}
            type="button"
            disabled={!ok}
            aria-pressed={active}
            onClick={() => applyPreset(p)}
            className={cn(
              'no-drag relative inline-flex h-8 items-center gap-1.5 rounded-xl border px-3 text-xs font-semibold transition-colors',
              active ? 'border-transparent text-accent-fg' : 'bg-elevated text-muted hover:text-fg dark:bg-line/[0.04]',
              !ok && 'cursor-not-allowed opacity-40',
            )}
          >
            {active && (
              <motion.span layoutId="preset-active" className="absolute inset-0 rounded-xl bg-accent-gradient shadow-glow" transition={{ type: 'spring', stiffness: 500, damping: 36 }} />
            )}
            <span className="relative flex items-center gap-1.5">
              {p === 'best' && <Sparkles className="size-3.5" aria-hidden />}
              {AUDIO.includes(p) && <Music className="size-3.5" aria-hidden />}
              {t(`result.presets.${p}`)}
            </span>
          </button>
        )
        return ok ? btn : <Tooltip key={p} content={t('result.presets.unavailable')}>{<span>{btn}</span>}</Tooltip>
      })}
      {info.isPlaylist && !info.matched && (
        <Tooltip content={t('result.video.playlistHint')}>
          <span className="text-subtle" tabIndex={0} aria-label={t('result.video.playlistHint')}>
            <Info className="size-3.5" />
          </span>
        </Tooltip>
      )}
    </div>
  )
}
