import { Download, History, House, Settings2, type LucideIcon } from 'lucide-react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import { useActiveCount } from '@/store/downloads'
import { useSettings } from '@/store/settings'
import { useUI, type Route } from '@/store/ui'
import { isWails } from '@/lib/api'
import { Tooltip } from '../ui/Tooltip'

const ITEMS: { route: Route; icon: LucideIcon; key: string }[] = [
  { route: 'home', icon: House, key: 'nav.home' },
  { route: 'downloads', icon: Download, key: 'nav.downloads' },
  { route: 'history', icon: History, key: 'nav.history' },
]

/** Slim navigation rail. It sits at the inline start, so it flips to the right in RTL. */
export function Sidebar() {
  const { t } = useTranslation()
  const route = useUI((s) => s.route)
  const navigate = useUI((s) => s.navigate)
  const active = useActiveCount()
  const fmt = useFormat()
  const mac = useSettings((s) => s.app?.os === 'darwin')

  const item = (r: Route, Icon: LucideIcon, label: string, badge?: number) => {
    const selected = route === r
    return (
      <Tooltip content={label} side={document.documentElement.dir === 'rtl' ? 'left' : 'right'} key={r}>
        <button
          type="button"
          onClick={() => navigate(r)}
          aria-current={selected ? 'page' : undefined}
          aria-label={badge ? `${label} — ${t('nav.activeBadge', { count: badge, n: fmt.number(badge) })}` : label}
          className={cn(
            'no-drag group relative flex w-[68px] flex-col items-center gap-1 rounded-xl px-1 py-2 text-2xs font-medium leading-tight transition-colors',
            selected ? 'text-fg' : 'text-subtle hover:text-fg',
          )}
        >
          {selected && (
            <motion.span
              layoutId="nav-active"
              className="absolute inset-0 rounded-xl bg-line/[0.07] dark:bg-line/[0.08]"
              transition={{ type: 'spring', stiffness: 500, damping: 40 }}
            />
          )}
          {selected && (
            <motion.span layoutId="nav-bar" className="absolute -start-2 top-[calc(50%-10px)] h-5 w-[3px] rounded-full bg-accent-gradient" />
          )}
          <span className="relative">
            <Icon className={cn('size-[19px] transition-transform group-active:scale-90', selected && 'text-accent-ink')} strokeWidth={selected ? 2.2 : 1.8} aria-hidden />
            {!!badge && (
              <motion.span
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className="absolute -end-2.5 -top-1.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-accent-gradient px-1 text-[10px] font-bold text-accent-fg"
              >
                {fmt.number(badge)}
              </motion.span>
            )}
          </span>
          <span className="relative line-clamp-2 max-w-full break-words text-center">{label}</span>
        </button>
      </Tooltip>
    )
  }

  return (
    <nav
      aria-label={t('nav.label')}
      className="drag glass relative z-20 flex w-[84px] shrink-0 flex-col items-center gap-1 border-e pb-3"
      style={{ paddingTop: mac ? 44 : 12 }}
    >
      {mac && !isWails && (
        <div className="pointer-events-none absolute start-[18px] top-[14px] flex gap-[7px]" aria-hidden>
          <span className="size-3 rounded-full bg-[#ff5f57]" />
          <span className="size-3 rounded-full bg-[#febc2e]" />
          <span className="size-3 rounded-full bg-[#28c840]" />
        </div>
      )}
      <div className="mb-3 flex size-9 items-center justify-center rounded-xl bg-accent-gradient shadow-glow" aria-hidden>
        <svg viewBox="0 0 24 24" className="size-[18px]" fill="none" stroke="rgb(var(--accent-fg))" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 4v10m0 0-4-4m4 4 4-4M5 14v4h14v-4" />
        </svg>
      </div>
      {ITEMS.map((i) => item(i.route, i.icon, t(i.key), i.route === 'downloads' ? active : undefined))}
      <div className="flex-1" />
      {item('settings', Settings2, t('nav.settings'))}
    </nav>
  )
}
