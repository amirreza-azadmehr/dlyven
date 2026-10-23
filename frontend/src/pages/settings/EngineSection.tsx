import { CheckCircle2, CircleAlert, Download, RefreshCw } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useFormat } from '@/lib/format'
import type { ComponentStatus } from '@/lib/types'
import { useEngine } from '@/store/engine'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { Button } from '@/components/ui/Button'
import { Badge, Switch } from '@/components/ui/Controls'
import { ProgressBar, Row, Section } from '@/components/ui/Feedback'
import { ErrorPanel } from '@/components/ErrorPanel'
import { PathField } from './DownloadsSection'

export function EngineSection() {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { settings: s, update } = useSettings()
  const engine = useEngine()
  const toast = useUI((x) => x.toast)
  const [checking, setChecking] = useState(false)
  if (!s) return null
  const st = engine.status

  const comp = (c: ComponentStatus | undefined, key: 'yt-dlp' | 'ffmpeg' | 'deno') => {
    const p = engine.progress[key]
    const busy = p && p.stage !== 'done' && p.stage !== 'error'
    return (
      <div key={key} className="rounded-xl border bg-line/[0.02] p-3.5">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2.5">
            {c?.installed ? <CheckCircle2 className="size-4 shrink-0 text-success" aria-hidden /> : <CircleAlert className="size-4 shrink-0 text-warning" aria-hidden />}
            <span className="text-[13px] font-semibold" dir="ltr">
              {t(`settings.component.${key}`)}
            </span>
            {c?.version && (
              <Badge>
                <bdi dir="ltr">{c.version}</bdi>
              </Badge>
            )}
            {c?.updateAvailable && <Badge tone="accent">{t('settings.updateAvailable', { version: c.latest })}</Badge>}
          </div>
          <span className="shrink-0 text-xs text-muted">{c?.installed ? t(`settings.source.${c.source}`) : t('settings.notInstalled')}</span>
        </div>
        {c?.path && (
          <bdi dir="ltr" className="mt-1.5 block truncate text-2xs text-subtle" title={c.path} data-selectable>
            {c.path}
          </bdi>
        )}
        {busy && (
          <div className="mt-2.5">
            <ProgressBar
              value={p.percent}
              indeterminate={p.stage !== 'downloading'}
              label={t(`settings.component.${key}`)}
              valueText={t(`engine.stages.${p.stage}`)}
            />
            <div className="mt-1 text-2xs text-muted">
              {t(`engine.stages.${p.stage}`)}
              {p.total > 0 && ` · ${fmt.bytes(p.downloaded)} / ${fmt.bytes(p.total)}`}
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <Section id="settings-engine" title={t('settings.sections.engine')} description={t('settings.engineIntro')}>
      <div className="grid gap-2.5">
        {comp(st?.ytDlp, 'yt-dlp')}
        {comp(st?.ffmpeg, 'ffmpeg')}
        {comp(st?.deno, 'deno')}
      </div>
      {engine.error && !engine.working && (
        <div className="mt-3">
          <ErrorPanel error={engine.error} />
        </div>
      )}
      <div className="mt-4 flex flex-wrap gap-2">
        {st && !st.ready ? (
          <Button variant="primary" loading={engine.working} icon={<Download className="size-4" aria-hidden />} onClick={() => void engine.install()}>
            {t('settings.installEngine')}
          </Button>
        ) : (
          <>
            <Button
              variant="primary"
              loading={engine.working}
              icon={<RefreshCw className="size-4" aria-hidden />}
              onClick={async () => {
                const ok = await engine.update(false)
                if (ok) toast({ kind: 'success', title: t('engine.updated') })
              }}
            >
              {t('settings.updateEngine')}
            </Button>
            <Button
              loading={checking}
              disabled={engine.working}
              onClick={async () => {
                setChecking(true)
                const res = await engine.check()
                setChecking(false)
                if (res && !res.ytDlp.updateAvailable) toast({ kind: 'info', title: t('engine.upToDate') })
              }}
            >
              {t('settings.checkUpdates')}
            </Button>
            <Button variant="ghost" disabled={engine.working} onClick={() => void engine.update(true)}>
              {t('settings.reinstall')}
            </Button>
          </>
        )}
      </div>

      <div className="mt-5">
        <Row label={t('settings.autoUpdate')} hint={t('settings.autoUpdateHint')}>
          <Switch label={t('settings.autoUpdate')} checked={s.autoUpdateEngine} onChange={(autoUpdateEngine) => update({ autoUpdateEngine })} />
        </Row>
        <Row label={t('settings.ytDlpPath')} stack>
          <PathField
            value={s.ytDlpPath}
            placeholder={t('settings.pathAuto')}
            onBrowse={async () => {
              const p = await api.ChooseExecutable(t('settings.chooseExecutableTitle')).catch(() => '')
              if (p) {
                update({ ytDlpPath: p })
                setTimeout(() => void engine.refresh(), 300)
              }
            }}
            onClear={() => {
              update({ ytDlpPath: '' })
              setTimeout(() => void engine.refresh(), 300)
            }}
          />
        </Row>
        <Row label={t('settings.ffmpegPath')} stack>
          <PathField
            value={s.ffmpegPath}
            placeholder={t('settings.pathAuto')}
            onBrowse={async () => {
              const p = await api.ChooseExecutable(t('settings.chooseExecutableTitle')).catch(() => '')
              if (p) {
                update({ ffmpegPath: p })
                setTimeout(() => void engine.refresh(), 300)
              }
            }}
            onClear={() => {
              update({ ffmpegPath: '' })
              setTimeout(() => void engine.refresh(), 300)
            }}
          />
        </Row>
      </div>
    </Section>
  )
}
