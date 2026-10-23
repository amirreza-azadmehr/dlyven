import * as RT from '@radix-ui/react-tabs'
import { AnimatePresence, motion } from 'framer-motion'
import { Captions, Image, Music, ShieldCheck, Video } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import type { MediaInfo } from '@/lib/types'
import { useAnalysis, type Tab } from '@/store/analysis'
import { AudioTab } from './AudioTab'
import { DownloadBar } from './DownloadBar'
import { PlaylistPanel } from './PlaylistPanel'
import { PresetsRow } from './Presets'
import { PreviewCard } from './PreviewCard'
import { SubtitlesTab } from './SubtitlesTab'
import { ThumbnailTab } from './ThumbnailTab'
import { VideoTab } from './VideoTab'

export function ResultView({ info }: { info: MediaInfo }) {
  const { t, i18n } = useTranslation()
  const { sel, select } = useAnalysis()
  const dir = i18n.dir() === 'rtl' ? -1 : 1

  const audioOnly = info.matched || info.platform.audioOnly
  const tabs: { id: Tab; icon: typeof Video; show: boolean }[] = [
    { id: 'video', icon: Video, show: !audioOnly },
    { id: 'audio', icon: Music, show: true },
    { id: 'subtitles', icon: Captions, show: !audioOnly && !info.isPlaylist },
    { id: 'thumbnail', icon: Image, show: !info.isPlaylist || info.matched },
  ]
  const visible = tabs.filter((x) => x.show)
  const order = visible.map((x) => x.id)

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
      <PreviewCard info={info} />

      {info.matched && (
        <motion.div initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} className="flex gap-3 rounded-2xl border border-success/25 bg-success/[0.06] p-4" role="note">
          <ShieldCheck className="mt-0.5 size-5 shrink-0 text-success" aria-hidden />
          <div>
            <div className="text-[13px] font-semibold">{t('result.spotify.title')}</div>
            <p className="mt-0.5 text-xs leading-relaxed text-muted">{t('result.spotify.body')}</p>
          </div>
        </motion.div>
      )}

      {info.isPlaylist && <PlaylistPanel info={info} />}

      <PresetsRow info={info} />

      <RT.Root value={sel.tab} onValueChange={(v) => select({ tab: v as Tab })} dir={i18n.dir()} className="card overflow-hidden">
        <RT.List aria-label={t('result.tabs.label')} className="flex gap-1 border-b p-1.5">
          {visible.map(({ id, icon: Icon }) => {
            const active = sel.tab === id
            return (
              <RT.Trigger
                key={id}
                value={id}
                className={cn('relative flex h-9 items-center gap-2 rounded-xl px-3.5 text-[13px] font-medium transition-colors', active ? 'text-fg' : 'text-muted hover:text-fg')}
              >
                {active && <motion.span layoutId="tab-pill" className="absolute inset-0 rounded-xl bg-line/[0.07]" transition={{ type: 'spring', stiffness: 500, damping: 38 }} />}
                <Icon className={cn('relative size-4', active && 'text-accent-ink')} aria-hidden />
                <span className="relative">{t(`result.tabs.${id}`)}</span>
              </RT.Trigger>
            )
          })}
        </RT.List>
        <AnimatePresence mode="wait" initial={false} custom={order.indexOf(sel.tab)}>
          <motion.div
            key={sel.tab}
            initial={{ opacity: 0, x: 14 * dir }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -14 * dir }}
            transition={{ duration: 0.16 }}
          >
            <RT.Content value={sel.tab} forceMount tabIndex={-1} className="outline-none">
              {sel.tab === 'video' && <VideoTab info={info} />}
              {sel.tab === 'audio' && <AudioTab info={info} />}
              {sel.tab === 'subtitles' && <SubtitlesTab info={info} />}
              {sel.tab === 'thumbnail' && <ThumbnailTab info={info} />}
            </RT.Content>
          </motion.div>
        </AnimatePresence>
      </RT.Root>

      <DownloadBar info={info} />
    </motion.div>
  )
}
