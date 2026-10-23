import { AnimatePresence } from 'framer-motion'
import { Link2, Trash2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useFormat } from '@/lib/format'
import { isFinal, useDownloads } from '@/store/downloads'
import { useUI } from '@/store/ui'
import { EmptyState, QueueIllustration } from '@/components/Illustrations'
import { Button } from '@/components/ui/Button'
import { Segmented } from '@/components/ui/Controls'
import { PageHeader } from '@/components/layout/PageHeader'
import { DownloadItem } from './DownloadItem'

type Filter = 'all' | 'active' | 'completed' | 'failed'

export function DownloadsPage() {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { jobs, order } = useDownloads()
  const navigate = useUI((s) => s.navigate)
  const [filter, setFilter] = useState<Filter>('all')

  const list = useMemo(() => {
    // Newest first, but keep active/queued items above finished ones.
    const all = order.map((id) => jobs[id]).filter(Boolean).reverse()
    const ranked = [...all.filter((j) => !isFinal(j)), ...all.filter(isFinal)]
    return ranked.filter((j) => {
      switch (filter) {
        case 'active':
          return !isFinal(j)
        case 'completed':
          return j.status === 'done'
        case 'failed':
          return j.status === 'failed' || j.status === 'canceled'
        default:
          return true
      }
    })
  }, [jobs, order, filter])

  const counts = useMemo(() => {
    const all = order.map((id) => jobs[id]).filter(Boolean)
    return {
      all: all.length,
      active: all.filter((j) => !isFinal(j)).length,
      completed: all.filter((j) => j.status === 'done').length,
      failed: all.filter((j) => j.status === 'failed' || j.status === 'canceled').length,
    }
  }, [jobs, order])

  const label = (f: Filter) => (
    <span className="flex items-center gap-1.5">
      {t(`downloads.filters.${f}`)}
      {counts[f] > 0 && <span className="tabular text-2xs text-subtle">{fmt.number(counts[f])}</span>}
    </span>
  )

  return (
    <div className="mx-auto w-full max-w-[920px] px-8 pb-10">
      <PageHeader
        title={t('downloads.title')}
        subtitle={t('downloads.subtitle')}
        actions={
          counts.all - counts.active > 0 && (
            <Button size="sm" variant="ghost" icon={<Trash2 className="size-3.5" aria-hidden />} onClick={() => void api.ClearFinishedDownloads()}>
              {t('downloads.clearFinished')}
            </Button>
          )
        }
      />
      {counts.all > 0 && (
        <Segmented<Filter>
          className="mb-4"
          label={t('downloads.filters.label')}
          value={filter}
          onChange={setFilter}
          options={(['all', 'active', 'completed', 'failed'] as Filter[]).map((f) => ({ value: f, label: label(f) }))}
        />
      )}
      {counts.all === 0 ? (
        <EmptyState
          art={<QueueIllustration />}
          title={t('downloads.emptyTitle')}
          body={t('downloads.emptyBody')}
          action={
            <Button variant="primary" icon={<Link2 className="size-4" aria-hidden />} onClick={() => navigate('home')}>
              {t('downloads.goHome')}
            </Button>
          }
        />
      ) : list.length === 0 ? (
        <p className="py-16 text-center text-[13px] text-muted">{t('downloads.emptyFilter')}</p>
      ) : (
        <ul className="space-y-2.5" aria-live="off">
          <AnimatePresence initial={false}>
            {list.map((j) => (
              <DownloadItem key={j.id} job={j} />
            ))}
          </AnimatePresence>
        </ul>
      )}
    </div>
  )
}
