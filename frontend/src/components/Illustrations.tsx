// Small inline SVG illustrations for empty states. They use theme tokens so they adapt to
// light/dark and the accent color. Decorative only (aria-hidden).
import type { ReactNode } from 'react'
import { motion } from 'framer-motion'

const accent = 'rgb(var(--accent))'
const accent2 = 'rgb(var(--accent-2))'
const line = 'rgb(var(--line) / 0.14)'
const fill = 'rgb(var(--line) / 0.05)'

export function QueueIllustration() {
  return (
    <svg width="168" height="120" viewBox="0 0 168 120" fill="none" aria-hidden>
      <defs>
        <linearGradient id="ig-q" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor={accent} />
          <stop offset="1" stopColor={accent2} />
        </linearGradient>
      </defs>
      <rect x="20" y="18" width="128" height="26" rx="8" fill={fill} stroke={line} />
      <rect x="30" y="28" width="56" height="6" rx="3" fill={line} />
      <rect x="20" y="52" width="128" height="26" rx="8" fill={fill} stroke={line} />
      <rect x="30" y="62" width="80" height="6" rx="3" fill={line} />
      <rect x="20" y="86" width="128" height="26" rx="8" fill={fill} stroke={line} opacity=".6" />
      <circle cx="132" cy="31" r="14" fill="url(#ig-q)" />
      <path d="M132 24v12m0 0-5-5m5 5 5-5" stroke="white" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export function HistoryIllustration() {
  return (
    <svg width="168" height="120" viewBox="0 0 168 120" fill="none" aria-hidden>
      <defs>
        <linearGradient id="ig-h" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor={accent} />
          <stop offset="1" stopColor={accent2} />
        </linearGradient>
      </defs>
      <circle cx="84" cy="60" r="40" fill={fill} stroke={line} />
      <circle cx="84" cy="60" r="28" stroke="url(#ig-h)" strokeWidth="3" strokeDasharray="4 6" strokeLinecap="round" />
      <path d="M84 44v17l10 7" stroke="url(#ig-h)" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round" />
      <circle cx="132" cy="28" r="5" fill={accent2} opacity=".5" />
      <circle cx="34" cy="92" r="4" fill={accent} opacity=".45" />
    </svg>
  )
}

export function SearchIllustration() {
  return (
    <svg width="140" height="100" viewBox="0 0 140 100" fill="none" aria-hidden>
      <circle cx="62" cy="46" r="26" fill={fill} stroke={line} strokeWidth="2" />
      <path d="m82 66 18 18" stroke={accent} strokeWidth="5" strokeLinecap="round" />
      <path d="M52 46h20" stroke={line} strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}

export function EmptyState({ art, title, body, action }: { art: ReactNode; title: string; body?: string; action?: ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex flex-col items-center justify-center px-6 py-16 text-center"
    >
      {art}
      <h3 className="mt-5 text-[15px] font-semibold text-fg">{title}</h3>
      {body && <p className="mt-1.5 max-w-sm text-[13px] text-muted">{body}</p>}
      {action && <div className="mt-5">{action}</div>}
    </motion.div>
  )
}
