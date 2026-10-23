import { Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useFormat } from '@/lib/format'
import type { MediaInfo, SubtitleTrack } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { Badge, Checkbox, Input, Segmented } from '@/components/ui/Controls'

const keyOf = (s: SubtitleTrack) => `${s.lang}|${s.auto ? 'a' : 'm'}`

function languageName(code: string, uiLang: string): string {
  try {
    return new Intl.DisplayNames([uiLang], { type: 'language' }).of(code.replace(/-orig$/, '')) ?? code
  } catch {
    return code
  }
}

export function SubtitlesTab({ info }: { info: MediaInfo }) {
  const { t, i18n } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const [q, setQ] = useState('')
  const chosen = new Set(sel.subs)

  const tracks = useMemo(
    () => info.subtitles.map((s) => ({ ...s, display: languageName(s.lang, i18n.language) })),
    [info.subtitles, i18n.language],
  )
  const filtered = tracks.filter((s) => {
    const n = q.trim().toLowerCase()
    return !n || s.display.toLowerCase().includes(n) || s.lang.toLowerCase().includes(n) || s.name.toLowerCase().includes(n)
  })
  const manual = filtered.filter((s) => !s.auto)
  const auto = filtered.filter((s) => s.auto)

  if (info.subtitles.length === 0) {
    return <p className="p-6 text-center text-[13px] text-muted">{t('result.subtitles.none')}</p>
  }

  const toggle = (k: string) => {
    const next = new Set(chosen)
    if (next.has(k)) next.delete(k)
    else next.add(k)
    select({ subs: [...next] })
  }

  const list = (title: string, items: typeof tracks) =>
    items.length > 0 && (
      <div className="mb-3">
        <div className="px-1 pb-1.5 text-xs font-semibold text-subtle">{title}</div>
        <div className="grid grid-cols-2 gap-1 sm:grid-cols-3">
          {items.map((s) => (
            <label key={keyOf(s)} className="flex cursor-pointer items-center gap-2.5 rounded-lg px-2 py-1.5 text-[13px] hover:bg-line/[0.05]">
              <Checkbox label={s.display} checked={chosen.has(keyOf(s))} onChange={() => toggle(keyOf(s))} />
              <span className="min-w-0 flex-1 truncate">
                <bdi>{s.display}</bdi>
                <span className="ms-1.5 text-2xs text-subtle" dir="ltr">
                  {s.lang}
                </span>
              </span>
              {s.auto && <Badge>{t('result.subtitles.autoBadge')}</Badge>}
            </label>
          ))}
        </div>
      </div>
    )

  return (
    <div>
      <div className="flex flex-wrap items-center gap-3 border-b p-3">
        <Input
          className="min-w-[200px] flex-1"
          icon={<Search className="size-4" />}
          placeholder={t('result.subtitles.search')}
          aria-label={t('result.subtitles.search')}
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <span className="text-xs text-muted" aria-live="polite">
          {t('result.subtitles.selected', { count: chosen.size, n: fmt.number(chosen.size) })}
        </span>
      </div>
      <div className="max-h-[300px] overflow-y-auto p-3">
        {list(t('result.subtitles.manual'), manual)}
        {list(t('result.subtitles.auto'), auto)}
        {filtered.length === 0 && <p className="py-6 text-center text-[13px] text-muted">{t('result.subtitles.noMatch', { q })}</p>}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
        <div className="flex flex-col gap-1.5">
          <span className="text-xs font-medium text-muted">{t('result.subtitles.mode')}</span>
          <Segmented
            size="sm"
            label={t('result.subtitles.mode')}
            value={sel.subMode}
            onChange={(subMode) => select({ subMode })}
            options={[
              { value: 'embed', label: t('result.subtitles.embed') },
              { value: 'file', label: t('result.subtitles.file') },
            ]}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <span className="text-xs font-medium text-muted">{t('result.subtitles.format')}</span>
          <Segmented
            size="sm"
            label={t('result.subtitles.format')}
            value={sel.subFormat}
            onChange={(subFormat) => select({ subFormat })}
            options={[
              { value: 'srt', label: 'SRT' },
              { value: 'vtt', label: 'VTT' },
            ]}
          />
        </div>
      </div>
      <p className="border-t px-4 py-2.5 text-xs leading-relaxed text-muted">{t('result.subtitles.hintEmbed')}</p>
    </div>
  )
}
