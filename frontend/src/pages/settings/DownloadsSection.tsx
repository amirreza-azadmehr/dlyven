import { FolderOpen, Minus, Plus, X } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useFormat } from '@/lib/format'
import type { AudioBitrate, AudioFormat, Container } from '@/lib/types'
import { useSettings } from '@/store/settings'
import { Button, IconButton } from '@/components/ui/Button'
import { Input, Segmented, Switch } from '@/components/ui/Controls'
import { Row, Section } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'

export function Stepper({ value, onChange, min, max, label }: { value: number; onChange(v: number): void; min: number; max: number; label: string }) {
  const fmt = useFormat()
  return (
    <div className="inline-flex h-9 items-center rounded-xl border bg-elevated dark:bg-line/[0.04]" role="group" aria-label={label}>
      <IconButton label={`${label} −`} size="sm" tooltip={false} disabled={value <= min} onClick={() => onChange(Math.max(min, value - 1))}>
        <Minus className="size-3.5" />
      </IconButton>
      <output className="tabular w-8 text-center text-[13px] font-semibold" aria-live="polite">
        {fmt.number(value)}
      </output>
      <IconButton label={`${label} +`} size="sm" tooltip={false} disabled={value >= max} onClick={() => onChange(Math.min(max, value + 1))}>
        <Plus className="size-3.5" />
      </IconButton>
    </div>
  )
}

export function PathField({ value, placeholder, onBrowse, onClear }: { value: string; placeholder?: string; onBrowse(): void; onClear?: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex w-full min-w-0 items-center gap-2">
      <div className="flex h-9 min-w-0 flex-1 items-center rounded-xl border bg-line/[0.03] px-3 text-[13px]">
        <bdi dir="ltr" className={value ? 'truncate text-fg' : 'truncate text-subtle'} title={value} data-selectable>
          {value || placeholder}
        </bdi>
      </div>
      {onClear && value && (
        <IconButton label={t('common.clear')} onClick={onClear}>
          <X className="size-4" />
        </IconButton>
      )}
      <Button size="md" icon={<FolderOpen className="size-4" aria-hidden />} onClick={onBrowse}>
        {t('common.browse')}
      </Button>
    </div>
  )
}

export function DownloadsSection() {
  const { t } = useTranslation()
  const { settings: s, update, app } = useSettings()
  const [preview, setPreview] = useState('')
  const [addFor, setAddFor] = useState('')
  const tplRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!s) return
    const vars: Record<string, string> = {
      title: 'Northern Lights — Timelapse',
      uploader: 'Aurora Films',
      quality: '1080p60',
      resolution: '1080p',
      fps: '60fps',
      platform: 'YouTube',
      id: 'dQw4w9WgXcQ',
      date: '2024-03-14',
      playlist: 'Nature',
      index: '03',
      artist: 'Nova Atlas',
      album: 'Tidal Lines',
      track: '07',
    }
    const id = setTimeout(() => void api.PreviewFilename(s.filenameTemplate, vars, s.defaultContainer).then(setPreview), 80)
    return () => clearTimeout(id)
  }, [s?.filenameTemplate, s?.defaultContainer])

  if (!s) return null

  const browse = async (current: string) => api.ChooseFolder(current, t('result.destination.chooseFolderTitle')).catch(() => '')

  const insertField = (f: string) => {
    const el = tplRef.current
    const token = `{${f}}`
    if (!el) return update({ filenameTemplate: s.filenameTemplate + token })
    const start = el.selectionStart ?? s.filenameTemplate.length
    const end = el.selectionEnd ?? start
    const next = s.filenameTemplate.slice(0, start) + token + s.filenameTemplate.slice(end)
    update({ filenameTemplate: next }, { debounce: true })
    requestAnimationFrame(() => {
      el.focus()
      el.setSelectionRange(start + token.length, start + token.length)
    })
  }

  const platforms = app?.platforms ?? []
  const unused = platforms.filter((p) => !s.platformDirs[p.id])

  return (
    <Section id="settings-downloads" title={t('settings.sections.downloads')}>
      <Row label={t('settings.downloadFolder')} stack>
        <PathField
          value={s.downloadDir}
          onBrowse={async () => {
            const d = await browse(s.downloadDir)
            if (d) update({ downloadDir: d })
          }}
        />
      </Row>
      <Row label={t('settings.rememberLastFolder')}>
        <Switch label={t('settings.rememberLastFolder')} checked={s.rememberLastFolder} onChange={(rememberLastFolder) => update({ rememberLastFolder })} />
      </Row>
      <Row label={t('settings.platformFolders')} hint={t('settings.platformFoldersHint')} stack>
        <div className="w-full space-y-2">
          {Object.entries(s.platformDirs).map(([pid, dir]) => (
            <div key={pid} className="flex items-center gap-2">
              <span className="w-28 shrink-0 truncate text-[13px] font-medium">
                <bdi>{platforms.find((p) => p.id === pid)?.name ?? pid}</bdi>
              </span>
              <PathField
                value={dir}
                onBrowse={async () => {
                  const d = await browse(dir)
                  if (d) update({ platformDirs: { ...s.platformDirs, [pid]: d } })
                }}
                onClear={() => {
                  const next = { ...s.platformDirs }
                  delete next[pid]
                  update({ platformDirs: next })
                }}
              />
            </div>
          ))}
          {unused.length > 0 && (
            <Select
              className="w-60"
              label={t('settings.addPlatformFolder')}
              value={addFor}
              onChange={async (pid) => {
                setAddFor('')
                if (!pid) return
                const d = await browse(s.downloadDir)
                if (d) update({ platformDirs: { ...s.platformDirs, [pid]: d } })
              }}
              options={[{ value: '', label: t('settings.addPlatformFolder') }, ...unused.map((p) => ({ value: p.id, label: p.name }))]}
            />
          )}
        </div>
      </Row>
      <Row label={t('settings.filenameTemplate')} hint={t('settings.filenameTemplateHint')} stack>
        <div className="w-full">
          <Input
            ref={tplRef}
            dir="ltr"
            spellCheck={false}
            value={s.filenameTemplate}
            aria-label={t('settings.filenameTemplate')}
            onChange={(e) => update({ filenameTemplate: e.target.value }, { debounce: true })}
          />
          <div className="mt-2 flex flex-wrap gap-1.5">
            {(app?.templateFields ?? []).map((f) => (
              <button
                key={f}
                type="button"
                onClick={() => insertField(f)}
                className="rounded-lg border bg-line/[0.03] px-2 py-1 text-xs text-muted transition-colors hover:border-accent/40 hover:text-fg"
                title={`{${f}}`}
              >
                {t(`settings.fields.${f}`)}
              </button>
            ))}
          </div>
          <div className="mt-3 rounded-xl bg-line/[0.04] px-3 py-2 text-xs">
            <span className="text-subtle">{t('settings.preview')}: </span>
            <bdi className="font-medium text-fg" dir="auto">
              {preview}
            </bdi>
          </div>
        </div>
      </Row>
      <Row label={t('settings.maxConcurrent')}>
        <Stepper label={t('settings.maxConcurrent')} value={s.maxConcurrent} min={1} max={10} onChange={(maxConcurrent) => update({ maxConcurrent })} />
      </Row>
      <Row label={t('settings.speedLimit')} hint={t('settings.speedLimitHint')}>
        <Input
          type="number"
          min={0}
          step={100}
          dir="ltr"
          className="w-32"
          aria-label={t('settings.speedLimit')}
          value={s.speedLimitKBps || ''}
          placeholder={t('common.unlimited')}
          onChange={(e) => update({ speedLimitKBps: Math.max(0, Number(e.target.value) || 0) }, { debounce: true })}
        />
      </Row>
      <Row label={t('settings.defaultContainer')}>
        <Segmented<Container>
          label={t('settings.defaultContainer')}
          value={s.defaultContainer}
          onChange={(defaultContainer) => update({ defaultContainer })}
          options={[
            { value: 'mp4', label: 'MP4' },
            { value: 'mkv', label: 'MKV' },
            { value: 'webm', label: 'WebM' },
          ]}
        />
      </Row>
      <Row label={t('settings.defaultAudio')}>
        <Select<AudioFormat>
          className="w-36"
          label={t('settings.defaultAudio')}
          value={s.defaultAudioFormat}
          onChange={(defaultAudioFormat) => update({ defaultAudioFormat })}
          options={(['mp3', 'm4a', 'opus', 'flac', 'wav'] as AudioFormat[]).map((f) => ({ value: f, label: f.toUpperCase(), hint: t(`result.audio.formats.${f}`) }))}
        />
      </Row>
      <Row label={t('settings.defaultBitrate')}>
        <Select<AudioBitrate>
          className="w-36"
          label={t('settings.defaultBitrate')}
          value={s.defaultAudioBitrate}
          onChange={(defaultAudioBitrate) => update({ defaultAudioBitrate })}
          options={(['best', '320', '256', '192', '128', '96'] as AudioBitrate[]).map((b) => ({ value: b, label: b === 'best' ? t('result.audio.best') : `${b} kbps` }))}
        />
      </Row>
      <Row label={t('settings.embedThumbnail')}>
        <Switch label={t('settings.embedThumbnail')} checked={s.embedThumbnail} onChange={(embedThumbnail) => update({ embedThumbnail })} />
      </Row>
      <Row label={t('settings.embedMetadata')}>
        <Switch label={t('settings.embedMetadata')} checked={s.embedMetadata} onChange={(embedMetadata) => update({ embedMetadata })} />
      </Row>
      <Row label={t('settings.embedChapters')}>
        <Switch label={t('settings.embedChapters')} checked={s.embedChapters} onChange={(embedChapters) => update({ embedChapters })} />
      </Row>
    </Section>
  )
}
