import { CheckCircle2, CircleArrowUp, Download, ExternalLink, Loader2, RefreshCw, RotateCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { rt } from '@/lib/api'
import { useFormat } from '@/lib/format'
import { useActiveCount } from '@/store/downloads'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { useUpdate } from '@/store/update'
import { Button } from '@/components/ui/Button'
import { Badge, Switch } from '@/components/ui/Controls'
import { ProgressBar, Row, Section } from '@/components/ui/Feedback'
import { ErrorPanel } from '@/components/ErrorPanel'

/** Dlyven's own updates: check GitHub, download + verify, restart into the new version. */
export function UpdatesSection() {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { settings: s, update } = useSettings()
  const u = useUpdate()
  const toast = useUI((x) => x.toast)
  const active = useActiveCount()
  if (!s) return null
  const st = u.status
  const p = u.progress
  const ready = !!st?.ready
  const busy = u.downloading || !!st?.busy
  const manual = st?.mode === 'manual'
  const openRelease = () => st && rt.BrowserOpenURL(st.releaseUrl)

  const title = ready
    ? t('updates.ready')
    : st?.available
      ? t('updates.available', { version: st.latest })
      : st?.latest
        ? t('updates.upToDate')
        : t('updates.notChecked')
  const body = ready
    ? st?.mode === 'installer'
      ? t('updates.readyBodyInstaller')
      : t('updates.readyBody')
    : st?.available
      ? manual
        ? t('updates.manualHint')
        : t('updates.availableBody')
      : null

  return (
    <Section id="settings-updates" title={t('settings.sections.updates')} description={t('updates.intro')}>
      <div className="rounded-xl border bg-line/[0.02] p-3.5" role="status" aria-live="polite">
        <div className="flex items-start gap-2.5">
          {busy || u.checking ? (
            <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin text-accent-ink" aria-hidden />
          ) : st?.available || ready ? (
            <CircleArrowUp className="mt-0.5 size-4 shrink-0 text-accent-ink" aria-hidden />
          ) : (
            <CheckCircle2 className="mt-0.5 size-4 shrink-0 text-success" aria-hidden />
          )}
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-[13px] font-semibold">{title}</span>
              {st?.current && (
                <Badge>
                  {t('updates.installed')} <bdi dir="ltr">{st.current}</bdi>
                </Badge>
              )}
            </div>
            {body && <p className="mt-1 text-xs leading-relaxed text-muted">{body}</p>}
            {ready && active > 0 && <p className="mt-1 text-xs leading-relaxed text-muted">{t('updates.activeDownloads')}</p>}
            {busy && p && p.stage !== 'ready' && p.stage !== 'error' && (
              <div className="mt-2.5">
                <ProgressBar
                  value={p.percent}
                  indeterminate={p.stage !== 'downloading' || !p.total}
                  label={t('updates.title')}
                  valueText={t(`updates.stages.${p.stage}`)}
                />
                <div className="mt-1 text-2xs text-muted">
                  {t(`updates.stages.${p.stage}`)}
                  {p.stage === 'downloading' && p.total > 0 && ` · ${fmt.bytes(p.downloaded)} / ${fmt.bytes(p.total)}`}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {u.error && !busy && (
        <div className="mt-3">
          <ErrorPanel
            error={u.error}
            showFix={u.error.code === 'ERR_NETWORK' || u.error.code === 'ERR_TIMEOUT' || u.error.code === 'ERR_PROXY'}
            onRetry={st?.available && !manual ? () => void u.download() : undefined}
          />
        </div>
      )}

      <div className="mt-4 flex flex-wrap gap-2">
        {ready ? (
          <Button variant="primary" icon={<RotateCw className="size-4" aria-hidden />} onClick={() => void u.restart()}>
            {t('updates.restart')}
          </Button>
        ) : st?.available && !manual ? (
          <Button variant="primary" loading={busy} icon={<Download className="size-4" aria-hidden />} onClick={() => void u.download()}>
            {t('updates.install')}
          </Button>
        ) : st?.available ? (
          <Button variant="primary" icon={<ExternalLink className="size-4" aria-hidden />} onClick={openRelease}>
            {t('updates.downloadManual')}
          </Button>
        ) : null}
        {!ready && (
          <Button
            loading={u.checking}
            disabled={busy}
            icon={<RefreshCw className="size-4" aria-hidden />}
            onClick={async () => {
              const res = await u.check()
              if (res && !res.available) toast({ kind: 'info', title: t('updates.upToDate') })
            }}
          >
            {t('updates.check')}
          </Button>
        )}
        {(st?.available || ready) && (
          <Button variant="ghost" icon={<ExternalLink className="size-4" aria-hidden />} onClick={openRelease}>
            {t('updates.whatsNew')}
          </Button>
        )}
        {u.error && st?.available && !manual && (
          <Button variant="ghost" onClick={openRelease}>
            {t('updates.downloadManual')}
          </Button>
        )}
      </div>

      <div className="mt-5">
        <Row label={t('updates.auto')} hint={t('updates.autoHint')}>
          <Switch label={t('updates.auto')} checked={s.autoCheckUpdates} onChange={(autoCheckUpdates) => update({ autoCheckUpdates })} />
        </Row>
      </div>
    </Section>
  )
}
