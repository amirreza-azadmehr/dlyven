import { Copy, Minus, Square, X } from 'lucide-react'
import { useEffect, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { CreatorCredit } from '@/components/CreatorCredit'
import { rt } from '@/lib/api'
import { cn } from '@/lib/cn'
import { useSettings } from '@/store/settings'

/**
 * Custom titlebar. The whole bar is a drag region (double-click maximizes — handled by Wails).
 * Windows/Linux: our own min/max/close buttons at the inline end, so they mirror in RTL.
 * macOS: native traffic lights (hidden-inset titlebar) stay top-left in every language,
 * so we reserve physical left space for them when the main column touches the left edge.
 */
export function TitleBar({ children }: { children?: ReactNode }) {
  const { t, i18n } = useTranslation()
  const os = useSettings((s) => s.app?.os)
  const [maximized, setMaximized] = useState(false)
  const mac = os === 'darwin'
  const rtl = i18n.dir() === 'rtl'

  useEffect(() => {
    if (mac) return
    const update = () => void rt.WindowIsMaximised().then(setMaximized).catch(() => {})
    update()
    window.addEventListener('resize', update)
    return () => window.removeEventListener('resize', update)
  }, [mac])

  return (
    <header
      className={cn('drag relative z-30 flex h-11 shrink-0 items-center gap-3 ps-4', mac && rtl && 'pl-[84px]')}
      style={mac ? undefined : { paddingInlineEnd: 0 }}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <CreatorCredit compact />
        {children}
      </div>
      {!mac && (
        <div className="no-drag flex h-full items-stretch" role="group" aria-label="Window">
          <WinButton label={t('titlebar.minimize')} onClick={() => rt.WindowMinimise()}>
            <Minus className="size-4" strokeWidth={1.6} />
          </WinButton>
          <WinButton
            label={maximized ? t('titlebar.restore') : t('titlebar.maximize')}
            onClick={() => {
              rt.WindowToggleMaximise()
              setTimeout(() => void rt.WindowIsMaximised().then(setMaximized), 150)
            }}
          >
            {maximized ? <Copy className="size-3.5 -scale-x-100" strokeWidth={1.6} /> : <Square className="size-3.5" strokeWidth={1.6} />}
          </WinButton>
          <WinButton label={t('titlebar.close')} onClick={() => rt.Quit()} danger>
            <X className="size-4" strokeWidth={1.6} />
          </WinButton>
        </div>
      )}
    </header>
  )
}

function WinButton({ label, onClick, children, danger }: { label: string; onClick(): void; children: ReactNode; danger?: boolean }) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      className={cn(
        'flex w-[46px] items-center justify-center text-muted transition-colors duration-100',
        danger ? 'hover:bg-[#E81123] hover:text-white' : 'hover:bg-line/[0.08] hover:text-fg',
      )}
    >
      {children}
    </button>
  )
}
