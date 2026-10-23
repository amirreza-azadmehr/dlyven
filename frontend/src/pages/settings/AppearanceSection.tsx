import { Check, Monitor, Moon, Sun } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { getLanguage, LANGUAGES } from '@/i18n'
import { cn } from '@/lib/cn'
import { createFormatter } from '@/lib/format'
import type { Accent, ThemeMode } from '@/lib/types'
import { useSettings } from '@/store/settings'
import { Segmented, Switch } from '@/components/ui/Controls'
import { Row, Section } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'

const ACCENTS: { id: Accent; from: string; to: string }[] = [
  { id: 'indigo', from: '#4F46E5', to: '#7C3AED' },
  { id: 'violet', from: '#7C3AED', to: '#C026D3' },
  { id: 'blue', from: '#2563EB', to: '#0891B2' },
  { id: 'teal', from: '#0F766E', to: '#047857' },
  { id: 'rose', from: '#E11D48', to: '#BE185D' },
  { id: 'amber', from: '#F59E0B', to: '#F97316' },
]

export function AppearanceSection() {
  const { t, i18n } = useTranslation()
  const { settings: s, update, app } = useSettings()
  if (!s) return null
  const lang = getLanguage(i18n.language)
  const sample = createFormatter(lang.code, true, 'gregorian').digits('0123')

  return (
    <Section id="settings-appearance" title={t('settings.sections.appearance')}>
      <Row label={t('settings.language')}>
        <Select
          className="w-56"
          label={t('settings.language')}
          value={s.language}
          onChange={(language) => update({ language })}
          options={[
            { value: '', label: t('settings.languageSystem') },
            ...LANGUAGES.map((l) => ({ value: l.code, label: l.nativeName, dir: l.dir })),
          ]}
        />
      </Row>
      <Row label={t('settings.theme')}>
        <Segmented<ThemeMode>
          label={t('settings.theme')}
          value={s.theme}
          onChange={(theme) => update({ theme })}
          options={[
            { value: 'system', label: <span className="flex items-center gap-1.5"><Monitor className="size-3.5" aria-hidden />{t('settings.themes.system')}</span> },
            { value: 'dark', label: <span className="flex items-center gap-1.5"><Moon className="size-3.5" aria-hidden />{t('settings.themes.dark')}</span> },
            { value: 'light', label: <span className="flex items-center gap-1.5"><Sun className="size-3.5" aria-hidden />{t('settings.themes.light')}</span> },
          ]}
        />
      </Row>
      <Row label={t('settings.accent')}>
        <div role="radiogroup" aria-label={t('settings.accent')} className="flex gap-2">
          {ACCENTS.map((a) => {
            const active = s.accent === a.id
            return (
              <button
                key={a.id}
                type="button"
                role="radio"
                aria-checked={active}
                aria-label={t(`settings.accents.${a.id}`)}
                title={t(`settings.accents.${a.id}`)}
                onClick={() => update({ accent: a.id })}
                className={cn('flex size-8 items-center justify-center rounded-full ring-offset-2 ring-offset-surface transition-transform hover:scale-110', active && 'ring-2 ring-fg/60')}
                style={{ backgroundImage: `linear-gradient(135deg, ${a.from}, ${a.to})` }}
              >
                {active && <Check className="size-4" style={{ color: a.id === 'amber' ? '#1c1406' : '#fff' }} strokeWidth={3} />}
              </button>
            )
          })}
        </div>
      </Row>
      {lang.nativeNumbering && (
        <Row label={t('settings.nativeDigits')} hint={t('settings.nativeDigitsHint', { sample })}>
          <Switch label={t('settings.nativeDigits')} checked={s.nativeDigits} onChange={(nativeDigits) => update({ nativeDigits })} />
        </Row>
      )}
      {lang.calendars?.includes('jalali') && (
        <Row label={t('settings.calendar')}>
          <Segmented
            label={t('settings.calendar')}
            value={s.calendar}
            onChange={(calendar) => update({ calendar })}
            options={[
              { value: 'gregorian', label: t('settings.calendars.gregorian') },
              { value: 'jalali', label: t('settings.calendars.jalali') },
            ]}
          />
        </Row>
      )}
      {app?.os !== 'linux' && (
        <Row label={t('settings.translucency')} hint={`${t('settings.translucencyHint')} ${t('common.restartRequired')}`}>
          <Switch label={t('settings.translucency')} checked={s.translucency} onChange={(translucency) => update({ translucency })} />
        </Row>
      )}
    </Section>
  )
}
