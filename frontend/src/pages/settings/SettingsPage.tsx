import { motion } from 'framer-motion'
import { CircleArrowUp, Cpu, Download, Globe2, KeyRound, Palette, SlidersHorizontal, Wrench } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useSettings } from '@/store/settings'
import { useUI, type SettingsSection } from '@/store/ui'
import { PageHeader } from '@/components/layout/PageHeader'
import { AppearanceSection } from './AppearanceSection'
import { DownloadsSection } from './DownloadsSection'
import { AdvancedSection, BehaviorSection, CookiesSection, NetworkSection } from './OtherSections'
import { EngineSection } from './EngineSection'
import { UpdatesSection } from './UpdatesSection'

const SECTIONS: { id: SettingsSection; icon: typeof Palette }[] = [
  { id: 'appearance', icon: Palette },
  { id: 'downloads', icon: Download },
  { id: 'network', icon: Globe2 },
  { id: 'cookies', icon: KeyRound },
  { id: 'engine', icon: Cpu },
  { id: 'updates', icon: CircleArrowUp },
  { id: 'behavior', icon: SlidersHorizontal },
  { id: 'advanced', icon: Wrench },
]

export function SettingsPage() {
  const { t } = useTranslation()
  const settings = useSettings((s) => s.settings)
  const target = useUI((s) => s.settingsSection)
  const scroller = useRef<HTMLDivElement>(null)
  const [current, setCurrent] = useState<SettingsSection>('appearance')

  const go = (id: SettingsSection) => {
    document.getElementById(`settings-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setCurrent(id)
  }

  useEffect(() => {
    if (target) setTimeout(() => go(target), 60)
  }, [target])

  // Highlight the section in view.
  useEffect(() => {
    const root = scroller.current?.closest('main')
    if (!root) return
    const obs = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0]
        if (visible) setCurrent(visible.target.id.replace('settings-', '') as SettingsSection)
      },
      { root, rootMargin: '-10% 0px -70% 0px' },
    )
    SECTIONS.forEach((s) => {
      const el = document.getElementById(`settings-${s.id}`)
      if (el) obs.observe(el)
    })
    return () => obs.disconnect()
  }, [settings === null])

  if (!settings) return null

  return (
    <div ref={scroller} className="mx-auto flex w-full max-w-[1040px] gap-8 px-8 pb-16">
      <aside className="sticky top-2 hidden h-fit w-48 shrink-0 pt-[76px] lg:block">
        <nav aria-label={t('settings.title')} className="space-y-0.5">
          {SECTIONS.map(({ id, icon: Icon }) => (
            <button
              key={id}
              type="button"
              onClick={() => go(id)}
              aria-current={current === id ? 'true' : undefined}
              className={cn('relative flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-start text-[13px] font-medium transition-colors', current === id ? 'text-fg' : 'text-muted hover:text-fg')}
            >
              {current === id && <motion.span layoutId="settings-nav" className="absolute inset-0 rounded-xl bg-line/[0.07]" transition={{ type: 'spring', stiffness: 500, damping: 40 }} />}
              <Icon className={cn('relative size-4', current === id && 'text-accent-ink')} aria-hidden />
              <span className="relative">{t(`settings.sections.${id}`)}</span>
            </button>
          ))}
        </nav>
      </aside>
      <div className="min-w-0 flex-1">
        <PageHeader title={t('settings.title')} subtitle={t('settings.subtitle')} />
        <div className="space-y-5">
          <AppearanceSection />
          <DownloadsSection />
          <NetworkSection />
          <CookiesSection />
          <EngineSection />
          <UpdatesSection />
          <BehaviorSection />
          <AdvancedSection />
        </div>
      </div>
    </div>
  )
}
