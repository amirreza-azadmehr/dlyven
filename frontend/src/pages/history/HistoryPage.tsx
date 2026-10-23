import { AnimatePresence, motion } from 'framer-motion'
import { AlertCircle, ExternalLink, FolderOpen, Music, RefreshCw, Search, Trash2 } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api, rt } from '@/lib/api'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { HistoryEntry } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { EmptyState, HistoryIllustration, SearchIllustration } from '@/components/Illustrations'
import { PlatformBadge } from '@/components/Platform'
import { Button, IconButton } from '@/components/ui/Button'
import { Badge, Input } from '@/components/ui/Controls'
import { ConfirmDialog } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'
import { PageHeader } from '@/components/layout/PageHeader'
import { errorTitle } from '@/components/ErrorPanel'
import { BidiList } from '@/components/ui/BidiList'

const PAGE = 50

export function HistoryPage() {
  const { t } = useTranslation()
  const fmt = useFormat()
  const registry = useSettings((s) => s.app?.platforms ?? [])
  const [search, setSearch] = useState('')
  const [platform, setPlatform] = useState('')
  const [status, setStatus] = useState('')
  const [items, setItems] = useState<HistoryEntry[]>([])
  const [total, setTotal] = useState(0)
  const [platforms, setPlatforms] = useState<string[]>([])
  const [missing, setMissing] = useState<Record<string, boolean>>({})
  const [confirm, setConfirm] = useState(false)
  const [loaded, setLoaded] = useState(false)

  const load = useCallback(
    async (offset = 0) => {
      try {
        const page = await api.ListHistory({ search, platform, status, limit: PAGE, offset })
        const next = Array.isArray(page?.items) ? page.items : []
        const n = typeof page?.total === 'number' ? page.total : next.length
        setItems((prev) => (offset ? [...prev, ...next] : next))
        setTotal(n)
        const checks = await Promise.all(next.filter((i) => i.status === 'done' && i.filePath).map(async (i) => [i.id, !(await api.FileExists(i.filePath))] as const))
        setMissing((m) => ({ ...m, ...Object.fromEntries(checks) }))
      } catch (e) {
        console.error('history load failed', e)
        if (!offset) {
          setItems([])
          setTotal(0)
        }
      } finally {
        setLoaded(true)
      }
    },
    [search, platform, status],
  )

  useEffect(() => {
    const id = setTimeout(() => void load(0), 180)
    return () => clearTimeout(id)
  }, [load])

  useEffect(() => {
    void api.HistoryPlatforms().then(setPlatforms)
    const off = rt.EventsOn('download:finished', () => void load(0))
    return off
  }, [load])

  const filtering = !!(search || platform || status)

  return (
    <div className="mx-auto w-full max-w-[920px] px-8 pb-10">
      <PageHeader
        title={t('history.title')}
        subtitle={t('history.subtitle')}
        actions={
          total > 0 && (
            <Button size="sm" variant="ghost" icon={<Trash2 className="size-3.5" aria-hidden />} onClick={() => setConfirm(true)}>
              {t('history.clear')}
            </Button>
          )
        }
      />

      {(total > 0 || filtering) && (
        <div className="mb-4 flex flex-wrap items-center gap-2">
          <Input
            className="min-w-[240px] flex-1"
            icon={<Search className="size-4" />}
            placeholder={t('history.search')}
            aria-label={t('history.search')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Select
            className="w-44"
            label={t('history.allPlatforms')}
            value={platform}
            onChange={setPlatform}
            options={[{ value: '', label: t('history.allPlatforms') }, ...platforms.map((p) => ({ value: p, label: registry.find((r) => r.id === p)?.name ?? p }))]}
          />
          <Select
            className="w-40"
            label={t('history.allStatuses')}
            value={status}
            onChange={setStatus}
            options={[
              { value: '', label: t('history.allStatuses') },
              { value: 'done', label: t('history.statusDone') },
              { value: 'failed', label: t('history.statusFailed') },
            ]}
          />
          <span className="tabular ms-auto text-xs text-subtle">{t('history.count', { count: total, n: fmt.number(total) })}</span>
        </div>
      )}

      {loaded && total === 0 ? (
        filtering ? (
          <EmptyState art={<SearchIllustration />} title={t('history.noResults')} />
        ) : (
          <EmptyState art={<HistoryIllustration />} title={t('history.emptyTitle')} body={t('history.emptyBody')} />
        )
      ) : (
        <ul className="divide-y overflow-hidden rounded-2xl border bg-surface shadow-soft">
          <AnimatePresence initial={false}>
            {items.map((h, i) => (
              <HistoryRow
                key={h.id}
                h={h}
                index={i}
                missing={!!missing[h.id]}
                platform={registry.find((p) => p.id === h.platform)}
                onDelete={async () => {
                  await api.DeleteHistory(h.id)
                  setItems((xs) => xs.filter((x) => x.id !== h.id))
                  setTotal((n) => n - 1)
                }}
              />
            ))}
          </AnimatePresence>
        </ul>
      )}
      {items.length < total && (
        <div className="mt-4 flex justify-center">
          <Button size="sm" onClick={() => void load(items.length)}>
            {t('history.loadMore')}
          </Button>
        </div>
      )}

      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t('history.clearTitle')}
        description={t('history.clearBody')}
        confirmLabel={t('history.clearConfirm')}
        danger
        onConfirm={async () => {
          await api.ClearHistory()
          setItems([])
          setTotal(0)
          setPlatforms([])
        }}
      />
    </div>
  )
}

function HistoryRow({
  h,
  index,
  missing,
  platform,
  onDelete,
}: {
  h: HistoryEntry
  index: number
  missing: boolean
  platform?: { id: string; name: string; color: string; icon: string }
  onDelete(): void
}) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const navigate = useUI((s) => s.navigate)
  const analyze = useAnalysis((s) => s.analyze)
  const square = h.kind === 'audio'
  return (
    <motion.li
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1, transition: { delay: Math.min(index, 15) * 0.015 } }}
      exit={{ opacity: 0, height: 0 }}
      className="group flex items-center gap-3.5 px-3.5 py-3 hover:bg-line/[0.03]"
    >
      <div className={cn('shrink-0 overflow-hidden rounded-lg bg-line/[0.06]', square ? 'size-11' : 'h-11 w-[72px]')}>
        {h.thumbnail ? (
          <img src={h.thumbnail} alt="" loading="lazy" referrerPolicy="no-referrer" className="size-full object-cover" />
        ) : (
          <div className="flex size-full items-center justify-center text-subtle">
            <Music className="size-4" />
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="truncate text-[13px] font-medium text-fg" dir="auto" title={h.title}>
          <bdi>{h.title || h.url}</bdi>
        </div>
        <div className="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-muted">
          {platform && <PlatformBadge platform={{ ...platform, known: true }} size="sm" className="h-5" />}
          {h.quality && (
            <Badge>
              <bdi dir="ltr">{h.quality}</bdi>
            </Badge>
          )}
          {h.status === 'failed' && <Badge tone="danger">{errorTitle(t, h.errorCode || 'ERR_UNKNOWN')}</Badge>}
          {missing && (
            <Badge tone="warning">
              <AlertCircle className="size-3" /> {t('history.fileMissing')}
            </Badge>
          )}
          <span className="tabular">
            <BidiList items={[h.size > 0 ? fmt.bytes(h.size) : '', fmt.dateTime(h.finishedAt)]} />
          </span>
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-0.5 opacity-70 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
        {h.status === 'done' && !missing && h.filePath && (
          <>
            <IconButton label={t('downloads.actions.openFile')} onClick={() => void api.OpenFile(h.filePath)}>
              <ExternalLink className="size-4 rtl:-scale-x-100" />
            </IconButton>
            <IconButton label={t('downloads.actions.showInFolder')} onClick={() => void api.ShowInFolder(h.filePath)}>
              <FolderOpen className="size-4" />
            </IconButton>
          </>
        )}
        {h.url && (
          <IconButton
            label={t('history.redownload')}
            onClick={() => {
              navigate('home')
              void analyze(h.url)
            }}
          >
            <RefreshCw className="size-4" />
          </IconButton>
        )}
        <IconButton label={t('history.delete')} tone="danger" onClick={onDelete}>
          <Trash2 className="size-4" />
        </IconButton>
      </div>
    </motion.li>
  )
}
