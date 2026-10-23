// Locale-aware formatting built on Intl. Every number shown in the UI goes through here so
// "native digits" (۰۱۲ for Persian, ٠١٢ for Kurdish) and the Jalali calendar apply everywhere.
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { getLanguage } from '@/i18n'
import { useSettings } from '@/store/settings'

export interface Formatter {
  locale: string
  number(n: number, opts?: Intl.NumberFormatOptions): string
  percent(p: number): string
  bytes(n: number, approx?: boolean): string
  speed(bytesPerSec: number): string
  duration(seconds: number): string
  date(isoOrMs: string | number, style?: 'medium' | 'long' | 'short'): string
  dateTime(ms: number): string
  compact(n: number): string
  bitrate(kbps: number): string
  digits(s: string | number): string
}

const UNITS = ['byte', 'kilobyte', 'megabyte', 'gigabyte', 'terabyte'] as const

export function createFormatter(lang: string, nativeDigits: boolean, calendar: 'gregorian' | 'jalali'): Formatter {
  const def = getLanguage(lang)
  const ext: string[] = []
  // Always explicit: fa-IR defaults to the Persian calendar in ICU, so "gregorian" must be forced.
  ext.push(calendar === 'jalali' && def.calendars?.includes('jalali') ? 'ca-persian' : 'ca-gregory')
  ext.push(`nu-${nativeDigits && def.nativeNumbering ? def.nativeNumbering : 'latn'}`)
  const locale = safeLocale(`${def.intlLocale}-u-${ext.join('-')}`, def.intlLocale)
  // Some engines ship without data for a locale (e.g. ckb in Chromium's trimmed ICU) and
  // silently fall back to English month names. Use a neutral numeric date there instead.
  const dateSupported = Intl.DateTimeFormat.supportedLocalesOf([def.intlLocale]).length > 0
  const nu = nativeDigits && def.nativeNumbering ? def.nativeNumbering : 'latn'
  const dateLocale = dateSupported ? locale : `en-CA-u-nu-${nu}` // en-CA → 2026-10-23
  // Values with Latin units ("12 MB", "320 kbps") are wrapped in Unicode isolates (FSI…PDI)
  // so they never reorder the surrounding words of an RTL sentence.
  const iso = (v: string) => (def.dir === 'rtl' ? `\u2068${v}\u2069` : v)
  const numericDate = (withTime: boolean): Intl.DateTimeFormatOptions => ({
    year: 'numeric', month: '2-digit', day: '2-digit', ...(withTime ? { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' as const } : {}),
  })

  const nf = (opts?: Intl.NumberFormatOptions) => new Intl.NumberFormat(locale, opts)
  const cache = new Map<string, Intl.NumberFormat>()
  const cached = (key: string, opts: Intl.NumberFormatOptions) => {
    let f = cache.get(key)
    if (!f) cache.set(key, (f = nf(opts)))
    return f
  }

  const bytes = (n: number, approx = false) => {
    if (!n || n < 0) return '—'
    let i = 0
    let v = n
    while (v >= 1024 && i < UNITS.length - 1) {
      v /= 1024
      i++
    }
    const digits = v >= 100 || i === 0 ? 0 : 1
    const s = cached(`b${i}${digits}`, { style: 'unit', unit: UNITS[i], unitDisplay: 'short', maximumFractionDigits: digits }).format(v)
    return iso(approx ? `~${s}` : s)
  }

  return {
    locale,
    number: (n, opts) => (opts ? nf(opts) : cached('n', {})).format(n),
    percent: (p) => cached('pct', { style: 'percent', maximumFractionDigits: 0 }).format(p / 100),
    bytes,
    speed: (bps) => (bps > 0 ? bytes(bps) : '—'),
    duration: (sec) => {
      if (!sec || sec < 0 || !isFinite(sec)) return '—'
      const s = Math.round(sec)
      const h = Math.floor(s / 3600)
      const m = Math.floor((s % 3600) / 60)
      const r = s % 60
      const two = cached('2d', { minimumIntegerDigits: 2, useGrouping: false })
      const one = cached('1d', { useGrouping: false })
      return h > 0 ? `${one.format(h)}:${two.format(m)}:${two.format(r)}` : `${one.format(m)}:${two.format(r)}`
    },
    date: (v, style = 'medium') => {
      if (!v) return ''
      const d = typeof v === 'number' ? new Date(v) : new Date(v.length === 10 ? v + 'T12:00:00' : v)
      if (isNaN(d.getTime())) return String(v)
      return iso(new Intl.DateTimeFormat(dateLocale, dateSupported ? { dateStyle: style } : numericDate(false)).format(d))
    },
    dateTime: (ms) =>
      iso(new Intl.DateTimeFormat(dateLocale, dateSupported ? { dateStyle: 'medium', timeStyle: 'short' } : numericDate(true)).format(new Date(ms))),
    compact: (n) => cached('c', { notation: 'compact', maximumFractionDigits: 1 }).format(n),
    bitrate: (kbps) => (kbps > 0 ? iso(`${cached('br', { maximumFractionDigits: 0 }).format(kbps)} kbps`) : '—'),
    digits: (s) => String(s).replace(/\d/g, (d) => cached('d', { useGrouping: false }).format(Number(d))),
  }
}

function safeLocale(tag: string, fallback: string): string {
  try {
    new Intl.NumberFormat(tag)
    return tag
  } catch {
    return fallback
  }
}

/** React hook returning a formatter bound to the current language & settings. */
export function useFormat(): Formatter {
  const { i18n } = useTranslation()
  const nativeDigits = useSettings((s) => s.settings?.nativeDigits ?? false)
  const calendar = useSettings((s) => s.settings?.calendar ?? 'gregorian')
  return useMemo(() => createFormatter(i18n.language, nativeDigits, calendar), [i18n.language, nativeDigits, calendar])
}
