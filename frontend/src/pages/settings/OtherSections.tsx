import { AlertTriangle, ClipboardCopy, FileText, FolderOpen, RotateCcw, ShieldAlert } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api, rt } from '@/lib/api'
import type { Settings } from '@/lib/types'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { APP_NAME } from '@/config/app'
import { CreatorCredit } from '@/components/CreatorCredit'
import { Button } from '@/components/ui/Button'
import { Input, Kbd, Segmented, Switch } from '@/components/ui/Controls'
import { ConfirmDialog, Row, Section } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'
import { PathField, Stepper } from './DownloadsSection'

export function NetworkSection() {
  const { t } = useTranslation()
  const { settings: s, update } = useSettings()
  if (!s) return null
  const on = s.proxyEnabled
  return (
    <Section id="settings-network" title={t('settings.sections.network')}>
      <Row label={t('settings.proxy')} hint={t('settings.proxyHint')}>
        <Switch label={t('settings.proxy')} checked={on} onChange={(proxyEnabled) => update({ proxyEnabled })} />
      </Row>
      {on && (
        <div className="grid grid-cols-[auto_1fr_110px] items-end gap-3 border-t py-3.5">
          <label className="flex flex-col gap-1.5 text-xs font-medium text-muted">
            {t('settings.proxyType')}
            <Segmented<Settings['proxyType']>
              size="sm"
              label={t('settings.proxyType')}
              value={s.proxyType}
              onChange={(proxyType) => update({ proxyType })}
              options={[
                { value: 'http', label: 'HTTP' },
                { value: 'https', label: 'HTTPS' },
                { value: 'socks5', label: 'SOCKS5' },
              ]}
            />
          </label>
          <label className="flex flex-col gap-1.5 text-xs font-medium text-muted">
            {t('settings.proxyHost')}
            <Input dir="ltr" spellCheck={false} placeholder={t('settings.proxyHostPlaceholder')} value={s.proxyHost} onChange={(e) => update({ proxyHost: e.target.value }, { debounce: true })} />
          </label>
          <label className="flex flex-col gap-1.5 text-xs font-medium text-muted">
            {t('settings.proxyPort')}
            <Input
              dir="ltr"
              type="number"
              min={0}
              max={65535}
              value={s.proxyPort || ''}
              onChange={(e) => update({ proxyPort: Math.max(0, Math.min(65535, Number(e.target.value) || 0)) }, { debounce: true })}
            />
          </label>
          <label className="col-span-3 grid grid-cols-2 gap-3 text-xs font-medium text-muted">
            <span className="flex flex-col gap-1.5">
              {t('settings.proxyUser')}
              <Input dir="ltr" autoComplete="off" placeholder={t('settings.optional')} value={s.proxyUser} onChange={(e) => update({ proxyUser: e.target.value }, { debounce: true })} />
            </span>
            <span className="flex flex-col gap-1.5">
              {t('settings.proxyPass')}
              <Input dir="ltr" type="password" autoComplete="new-password" placeholder={t('settings.optional')} value={s.proxyPass} onChange={(e) => update({ proxyPass: e.target.value }, { debounce: true })} />
            </span>
          </label>
        </div>
      )}
      <Row label={t('settings.userAgent')} hint={t('settings.userAgentHint')} stack>
        <Input dir="ltr" spellCheck={false} value={s.userAgent} placeholder="Mozilla/5.0 …" aria-label={t('settings.userAgent')} onChange={(e) => update({ userAgent: e.target.value }, { debounce: true })} className="w-full" />
      </Row>
      <Row label={t('settings.retries')} hint={t('settings.retriesHint')}>
        <Stepper label={t('settings.retries')} value={s.retries} min={0} max={10} onChange={(retries) => update({ retries })} />
      </Row>
      <Row label={t('settings.timeout')}>
        <div className="flex items-center gap-2">
          <Input
            dir="ltr"
            type="number"
            min={5}
            max={300}
            className="w-24"
            aria-label={t('settings.timeout')}
            value={s.timeoutSec}
            onChange={(e) => update({ timeoutSec: Number(e.target.value) || 30 }, { debounce: true })}
          />
          <span className="text-xs text-muted">{t('common.seconds')}</span>
        </div>
      </Row>
    </Section>
  )
}

export function CookiesSection() {
  const { t } = useTranslation()
  const { settings: s, update, app } = useSettings()
  if (!s) return null
  return (
    <Section id="settings-cookies" title={t('settings.sections.cookies')} description={t('settings.cookiesIntro')}>
      <Row label={t('settings.cookiesMode')}>
        <Select<Settings['cookiesMode']>
          className="w-64"
          label={t('settings.cookiesMode')}
          value={s.cookiesMode}
          onChange={(cookiesMode) => update({ cookiesMode })}
          options={(['none', 'browser', 'file'] as const).map((m) => ({ value: m, label: t(`settings.cookiesModes.${m}`) }))}
        />
      </Row>
      {s.cookiesMode === 'browser' && (
        <Row label={t('settings.cookiesBrowser')} hint={t('settings.cookiesBrowserHint')}>
          <Select
            className="w-44"
            label={t('settings.cookiesBrowser')}
            value={s.cookiesBrowser}
            onChange={(cookiesBrowser) => update({ cookiesBrowser })}
            options={(app?.cookieBrowsers ?? []).map((b) => ({ value: b, label: b.charAt(0).toUpperCase() + b.slice(1), dir: 'ltr' as const }))}
          />
        </Row>
      )}
      {s.cookiesMode === 'file' && (
        <Row label={t('settings.cookiesFile')} hint={t('settings.cookiesFileHint')} stack>
          <PathField
            value={s.cookiesFile}
            onBrowse={async () => {
              const f = await api.ChooseCookiesFile(t('settings.chooseCookies')).catch(() => '')
              if (f) update({ cookiesFile: f })
            }}
            onClear={() => update({ cookiesFile: '' })}
          />
        </Row>
      )}
      {s.cookiesMode !== 'none' && (
        <div className="mt-3 flex gap-2.5 rounded-xl border border-warning/25 bg-warning/[0.06] p-3 text-xs leading-relaxed text-muted" role="note">
          <AlertTriangle className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden />
          {t('settings.cookiesWarning')}
        </div>
      )}
    </Section>
  )
}

export function BehaviorSection() {
  const { t } = useTranslation()
  const { settings: s, update } = useSettings()
  if (!s) return null
  return (
    <Section id="settings-behavior" title={t('settings.sections.behavior')}>
      <Row label={t('settings.clipboardWatch')} hint={t('settings.clipboardWatchHint')}>
        <Switch label={t('settings.clipboardWatch')} checked={s.clipboardWatch} onChange={(clipboardWatch) => update({ clipboardWatch })} />
      </Row>
      <div className="border-t pt-3.5 text-xs font-semibold text-subtle">{t('settings.afterDownload')}</div>
      <Row label={t('settings.openFolderOnDone')}>
        <Switch label={t('settings.openFolderOnDone')} checked={s.openFolderOnDone} onChange={(openFolderOnDone) => update({ openFolderOnDone })} />
      </Row>
      <Row label={t('settings.notify')}>
        <Switch label={t('settings.notify')} checked={s.notify} onChange={(notify) => update({ notify })} />
      </Row>
      <Row label={t('settings.playSound')}>
        <Switch label={t('settings.playSound')} checked={s.playSound} onChange={(playSound) => update({ playSound })} />
      </Row>
    </Section>
  )
}

export function AdvancedSection() {
  const { t } = useTranslation()
  const { app, reset } = useSettings()
  const toast = useUI((x) => x.toast)
  const [confirm, setConfirm] = useState(false)
  const mod = app?.os === 'darwin' ? '⌘' : 'Ctrl'

  return (
    <Section id="settings-advanced" title={t('settings.sections.advanced')}>
      <Row label={t('settings.diagnostics')} hint={t('settings.diagnosticsHint')}>
        <Button
          icon={<ClipboardCopy className="size-4" aria-hidden />}
          onClick={async () => {
            const text = await api.GetDiagnostics()
            await rt.ClipboardSetText(text)
            toast({ kind: 'success', title: t('common.copied') })
          }}
        >
          {t('settings.copyDiagnostics')}
        </Button>
        <Button variant="ghost" icon={<FileText className="size-4" aria-hidden />} onClick={() => void api.OpenLogsFolder()}>
          {t('settings.openLogs')}
        </Button>
      </Row>
      <Row label={t('settings.dataFolder')}>
        <Button variant="ghost" icon={<FolderOpen className="size-4" aria-hidden />} onClick={() => app && void api.OpenFolder(app.dataDir)}>
          <bdi dir="ltr" className="max-w-[260px] truncate">
            {app?.dataDir}
          </bdi>
        </Button>
      </Row>
      <Row label={t('settings.shortcuts')} stack>
        <dl className="grid w-full grid-cols-2 gap-x-6 gap-y-2 text-[13px]">
          {[
            [`${mod} V`, 'paste'],
            [`${mod} ↵`, 'download'],
            [`${mod} ,`, 'settings'],
            ['Esc', 'cancel'],
          ].map(([k, id]) => (
            <div key={id} className="flex items-center justify-between gap-3">
              <dt className="text-muted">{t(`settings.shortcutList.${id}`)}</dt>
              <dd>
                <Kbd>{k}</Kbd>
              </dd>
            </div>
          ))}
        </dl>
      </Row>
      <Row label={t('settings.resetAll')}>
        <Button variant="danger" icon={<RotateCcw className="size-4" aria-hidden />} onClick={() => setConfirm(true)}>
          {t('settings.resetAll')}
        </Button>
      </Row>
      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-4 text-xs text-subtle">
        <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <span>{t('settings.about', { name: app?.name ?? APP_NAME, version: app?.version ?? '' })}</span>
          <CreatorCredit className="text-subtle hover:text-fg" />
        </span>
        <span className="flex items-center gap-1.5">
          <ShieldAlert className="size-3.5" aria-hidden />
          {t('app.legal')}
        </span>
      </div>
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t('settings.resetTitle')}
        description={t('settings.resetBody')}
        confirmLabel={t('settings.resetAll')}
        danger
        onConfirm={async () => {
          await reset()
          toast({ kind: 'success', title: t('settings.resetDone') })
        }}
      />
    </Section>
  )
}
