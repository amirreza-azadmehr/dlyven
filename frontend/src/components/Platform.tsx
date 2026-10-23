import {
  AtSign, Camera, Cloud, Disc3, Film, Gamepad2, Globe, MessageCircle, Music, Music2, Pin, Play, SquarePlay, Tv, Users, Video,
  type LucideIcon,
} from 'lucide-react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import type { PlatformInfo } from '@/lib/types'

// Generic glyphs only — no trademarked logos. Brand identity comes from the tint color.
const ICONS: Record<string, LucideIcon> = {
  'square-play': SquarePlay, camera: Camera, music: Music, 'music-2': Music2, 'at-sign': AtSign, users: Users,
  video: Video, cloud: Cloud, 'message-circle': MessageCircle, 'gamepad-2': Gamepad2, play: Play, 'disc-3': Disc3,
  pin: Pin, tv: Tv, film: Film, globe: Globe,
}

export function PlatformIcon({ icon, className }: { icon: string; className?: string }) {
  const I = ICONS[icon] ?? Globe
  return <I className={className} aria-hidden strokeWidth={2} />
}

export function PlatformBadge({ platform, size = 'md', className }: { platform: Pick<PlatformInfo, 'id' | 'name' | 'color' | 'icon' | 'known'>; size?: 'sm' | 'md'; className?: string }) {
  const { t } = useTranslation()
  const name = platform.known ? platform.name : platform.name || t('home.otherSite')
  return (
    <motion.span
      layout
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn('inline-flex max-w-[200px] items-center gap-1.5 rounded-full border font-medium text-fg', size === 'sm' ? 'h-6 px-2 text-2xs' : 'h-7 px-2.5 text-xs', className)}
      // 8-digit hex keeps this working on older WebKit; text stays theme-colored for contrast.
      style={{ backgroundColor: `${platform.color}1f`, borderColor: `${platform.color}40` }}
    >
      <span style={{ color: platform.color }} className="flex">
        <PlatformIcon icon={platform.icon} className={size === 'sm' ? 'size-3' : 'size-3.5'} />
      </span>
      <bdi className="truncate">{name}</bdi>
    </motion.span>
  )
}
