import { ArrowDownToLine, FileText, FolderOpen } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api, toAppError } from '@/lib/api'
import { useFormat } from '@/lib/format'
import { buildRequests, expectedExt, previewVars } from '@/lib/requests'
import type { MediaInfo } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { Button } from '@/components/ui/Button'
import { Kbd } from '@/components/ui/Controls'
import { Tooltip } from '@/components/ui/Tooltip'
import { errorHint, errorTitle } from '@/components/ErrorPanel'

export const DOWNLOAD_EVENT = 'dlyven:download'

export function DownloadBar({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const settings = useSettings((s) => s.settings)!
  const { toast, navigate } = useUI()
  const [preview, setPreview] = useState('')
  const [busy, setBusy] = useState(false)

  const count = info.isPlaylist ? sel.entries.length : 1
  const disabled = count === 0 || (sel.tab === 'subtitles' && sel.subs.length === 0) || (info.matched && count === 0)

  useEffect(() => {
    const id = setTimeout(() => {
      void api.PreviewFilename(settings.filenameTemplate, previewVars(info, sel), expectedExt(sel)).then(setPreview)
    }, 80)
    return () => clearTimeout(id)
  }, [info, sel, settings.filenameTemplate])

  const start = async () => {
    if (disabled || busy) {
      if (count === 0) toast({ kind: 'info', title: t('result.nothingSelected') })
      else if (sel.tab === 'subtitles' && sel.subs.length === 0) toast({ kind: 'info', title: t('result.noSubsSelected') })
      return
    }
    setBusy(true)
    try {
      const reqs = buildRequests(info, sel, settings)
      if (reqs.length === 1) await api.Enqueue(reqs[0])
      else await api.EnqueueMany(reqs)
      toast({
        kind: 'success',
        title: reqs.length === 1 ? t('result.queued') : t('result.queuedMany', { count: reqs.length, n: fmt.number(reqs.length) }),
        body: reqs.length === 1 ? reqs[0].title : undefined,
        action: { label: t('result.viewDownloads'), run: () => navigate('downloads') },
      })
    } catch (e) {
      const err = toAppError(e)
      toast({ kind: 'error', title: errorTitle(t, err.code), body: errorHint(t, err.code) })
    } finally {
      setBusy(false)
    }
  }

  useEffect(() => {
    const h = () => void start()
    window.addEventListener(DOWNLOAD_EVENT, h)
    return () => window.removeEventListener(DOWNLOAD_EVENT, h)
  })

  const chooseFolder = async () => {
    const dir = await api.ChooseFolder(sel.outputDir, t('result.destination.chooseFolderTitle')).catch(() => '')
    if (dir) select({ outputDir: dir })
  }

  const mac = navigator.platform.toLowerCase().includes('mac')

  return (
    <div className="glass sticky bottom-0 z-10 -mx-1 mt-4 flex items-center gap-4 rounded-2xl border p-3 shadow-lift">
      <div className="grid min-w-0 flex-1 grid-cols-[auto_1fr] items-center gap-x-3 gap-y-1.5 text-[13px]">
        <span className="flex items-center gap-1.5 text-xs text-muted">
          <FolderOpen className="size-3.5" aria-hidden /> {t('result.destination.folder')}
        </span>
        <div className="flex min-w-0 items-center gap-2">
          <bdi dir="ltr" className="min-w-0 truncate font-medium text-fg" title={sel.outputDir} data-selectable>
            {sel.outputDir}
          </bdi>
          <button type="button" onClick={() => void chooseFolder()} className="shrink-0 rounded-md px-1.5 text-xs font-semibold text-accent-ink hover:underline">
            {t('common.change')}
          </button>
        </div>
        <span className="flex items-center gap-1.5 text-xs text-muted">
          <FileText className="size-3.5" aria-hidden /> {count > 1 ? t('result.destination.filenameMulti') : t('result.destination.filename')}
        </span>
        <Tooltip content={t('result.destination.editTemplate')}>
          <button type="button" onClick={() => navigate('settings', 'downloads')} className="min-w-0 truncate text-start text-muted hover:text-fg" dir="auto">
            <bdi>{preview}</bdi>
          </button>
        </Tooltip>
      </div>
      <Button
        variant="primary"
        size="lg"
        loading={busy}
        disabled={disabled}
        onClick={() => void start()}
        icon={<ArrowDownToLine className="size-[18px]" aria-hidden />}
        className="min-w-[156px]"
      >
        {count > 1 ? t('result.downloadCount', { count, n: fmt.number(count) }) : t('result.download')}
        <span className="ms-1 hidden opacity-75 lg:inline-flex">
          <Kbd>{mac ? '⌘↵' : 'Ctrl ↵'}</Kbd>
        </span>
      </Button>
    </div>
  )
}
