import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api, toAppError } from '@/lib/api'
import { useFormat } from '@/lib/format'
import { detectPlatform } from '@/lib/platforms'
import { buildRequests } from '@/lib/requests'
import type { DownloadRequest } from '@/lib/types'
import { useAnalysis, type Selection } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { PlatformBadge } from '@/components/Platform'
import { Button } from '@/components/ui/Button'
import { Dialog, ProgressBar } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'

type Mode = 'video' | '1080' | 'mp3' | 'm4a'

/** Shown when several links are pasted or a .txt file is dropped. */
export function BatchDialog({ urls, onClose }: { urls: string[]; onClose(): void }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const settings = useSettings((s) => s.settings)!
  const registry = useSettings((s) => s.app?.platforms ?? [])
  const { toast, navigate } = useUI()
  const analyze = useAnalysis((s) => s.analyze)
  const [mode, setMode] = useState<Mode>('video')
  const [done, setDone] = useState(-1)
  const [failed, setFailed] = useState(0)
  const open = urls.length > 0
  const running = done >= 0

  useEffect(() => {
    if (!open) {
      setDone(-1)
      setFailed(0)
    }
  }, [open])

  const run = async () => {
    setDone(0)
    let queued = 0
    let errors = 0
    for (const [i, url] of urls.entries()) {
      try {
        const info = await api.AnalyzeForBatch(url)
        const sel: Selection = {
          tab: mode === 'mp3' || mode === 'm4a' || info.matched || info.platform.audioOnly ? 'audio' : 'video',
          formatId: null,
          maxHeight: mode === '1080' ? 1080 : 0,
          preset: null,
          container: settings.defaultContainer,
          audioFormat: mode === 'm4a' ? 'm4a' : mode === 'mp3' ? 'mp3' : settings.defaultAudioFormat,
          audioBitrate: mode === 'm4a' ? 'best' : settings.defaultAudioBitrate,
          audioFormatId: '',
          subs: [],
          subMode: 'embed',
          subFormat: 'srt',
          thumbUrl: '',
          thumbFormat: 'jpg',
          embedThumbnail: settings.embedThumbnail,
          embedMetadata: settings.embedMetadata,
          entries: info.entries.map((e) => e.index),
          outputDir: await api.DefaultFolderFor(info.platform.id),
        }
        const reqs: DownloadRequest[] = buildRequests(info, sel, settings)
        await api.EnqueueMany(reqs)
        queued += reqs.length
      } catch (e) {
        console.warn('batch item failed', url, toAppError(e))
        errors++
        setFailed(errors)
      }
      setDone(i + 1)
    }
    toast({
      kind: errors && !queued ? 'error' : 'success',
      title: t('batch.queued', { count: queued, n: fmt.number(queued) }),
      action: { label: t('result.viewDownloads'), run: () => navigate('downloads') },
    })
    onClose()
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => !v && !running && onClose()}
      title={t('batch.title', { count: urls.length, n: fmt.number(urls.length) })}
      description={t('batch.subtitle')}
      footer={
        <>
          <Button
            variant="ghost"
            disabled={running}
            onClick={() => {
              onClose()
              void analyze(urls[0])
            }}
          >
            {t('batch.analyzeFirst')}
          </Button>
          <Button variant="primary" loading={running} onClick={() => void run()}>
            {t('batch.start', { count: urls.length, n: fmt.number(urls.length) })}
          </Button>
        </>
      }
    >
      <ul className="max-h-56 space-y-1 overflow-y-auto rounded-xl border p-1.5">
        {urls.map((u) => (
          <li key={u} className="flex items-center gap-2 rounded-lg px-2 py-1.5">
            <PlatformBadge platform={detectPlatform(u, registry)} size="sm" />
            <bdi dir="ltr" className="min-w-0 flex-1 truncate text-xs text-muted">
              {u}
            </bdi>
          </li>
        ))}
      </ul>
      <div className="mt-4 flex items-center justify-between gap-3">
        <span className="text-[13px] font-medium">{t('batch.mode')}</span>
        <Select<Mode>
          className="w-56"
          label={t('batch.mode')}
          value={mode}
          onChange={setMode}
          disabled={running}
          options={[
            { value: 'video', label: t('batch.modeVideo') },
            { value: '1080', label: t('batch.mode1080') },
            { value: 'mp3', label: t('batch.modeMp3') },
            { value: 'm4a', label: t('batch.modeM4a') },
          ]}
        />
      </div>
      {running && (
        <div className="mt-4">
          <ProgressBar value={(done / urls.length) * 100} label={t('home.analyzing')} valueText={`${fmt.number(done)} / ${fmt.number(urls.length)}`} />
          <p className="mt-1.5 text-xs text-muted">
            {t('home.analyzing')} {fmt.number(done)} / {fmt.number(urls.length)}
            {failed > 0 && ` · ${t('downloads.status.failed')}: ${fmt.number(failed)}`}
          </p>
        </div>
      )}
    </Dialog>
  )
}
