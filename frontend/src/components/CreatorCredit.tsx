import { Github, Linkedin, Send } from 'lucide-react'
import { useState, type ComponentType } from 'react'
import { useTranslation } from 'react-i18next'
import { AUTHOR, type AuthorLinkId } from '@/config/author'
import { rt } from '@/lib/api'
import { cn } from '@/lib/cn'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Feedback'

const ICONS: Record<AuthorLinkId, ComponentType<{ className?: string }>> = {
  github: Github,
  linkedin: Linkedin,
  telegram: Send,
}

export function CreatorCredit({ className, compact }: { className?: string; compact?: boolean }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <>
      <button
        type="button"
        className={cn(
          'no-drag rounded-lg px-2 py-1 text-left text-xs text-muted transition-colors hover:bg-line/[0.07] hover:text-fg',
          compact && 'max-w-[220px] truncate',
          className,
        )}
        onClick={() => setOpen(true)}
      >
        {t('creator.by', { name: AUTHOR.name })}
      </button>
      <CreatorDialog open={open} onOpenChange={setOpen} />
    </>
  )
}

export function CreatorDialog({ open, onOpenChange }: { open: boolean; onOpenChange(v: boolean): void }) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={AUTHOR.name}
      description={t('creator.subtitle')}
      footer={
        <Button variant="ghost" onClick={() => onOpenChange(false)}>
          {t('common.close')}
        </Button>
      }
    >
      <ul className="flex flex-col gap-2">
        {AUTHOR.links.map((link) => {
          const Icon = ICONS[link.id]
          return (
            <li key={link.id}>
              <button
                type="button"
                dir="ltr"
                className="flex w-full items-center gap-3 rounded-xl border bg-line/[0.03] px-3 py-2.5 text-start transition-colors hover:bg-line/[0.08]"
                onClick={() => rt.BrowserOpenURL(link.href)}
              >
                <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-line/[0.08] text-fg">
                  <Icon className="size-4" aria-hidden />
                </span>
                <span className="min-w-0">
                  <span className="block text-[13px] font-medium text-fg">{t(`creator.links.${link.id}`)}</span>
                  <span className="block truncate text-xs text-muted">{link.display}</span>
                </span>
              </button>
            </li>
          )
        })}
      </ul>
    </Dialog>
  )
}
