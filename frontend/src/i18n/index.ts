import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from '../locales/en.json'
import fa from '../locales/fa.json'
import tr from '../locales/tr.json'
import ckb from '../locales/ckb.json'

/**
 * Language registry — the ONLY place to register a language.
 * To add one: drop `src/locales/<code>.json` (copy en.json), import it above and add an entry here.
 *  - nativeName: shown in the language picker (no flags — languages are not countries)
 *  - dir: text direction; the whole layout mirrors automatically for 'rtl'
 *  - intlLocale: BCP-47 tag used for Intl number/date formatting
 *  - nativeNumbering: numbering system used when "native digits" is enabled (optional)
 *  - calendars: calendars the user may pick (gregorian is always available)
 */
export interface LanguageDef {
  code: string
  nativeName: string
  dir: 'ltr' | 'rtl'
  intlLocale: string
  nativeNumbering?: string
  calendars?: Array<'gregorian' | 'jalali'>
  resources: Record<string, unknown>
}

export const LANGUAGES: LanguageDef[] = [
  { code: 'en', nativeName: 'English', dir: 'ltr', intlLocale: 'en-US', resources: en },
  { code: 'fa', nativeName: 'فارسی', dir: 'rtl', intlLocale: 'fa-IR', nativeNumbering: 'arabext', calendars: ['gregorian', 'jalali'], resources: fa },
  { code: 'tr', nativeName: 'Türkçe', dir: 'ltr', intlLocale: 'tr-TR', resources: tr },
  { code: 'ckb', nativeName: 'کوردی (سۆرانی)', dir: 'rtl', intlLocale: 'ckb-IQ', nativeNumbering: 'arab', resources: ckb },
]

export const FALLBACK = 'en'

export function getLanguage(code: string): LanguageDef {
  return LANGUAGES.find((l) => l.code === code) ?? LANGUAGES[0]
}

/** Picks the best supported language from the OS/browser preferences. */
export function detectSystemLanguage(): string {
  const prefs = typeof navigator !== 'undefined' ? navigator.languages ?? [navigator.language] : []
  for (const p of prefs) {
    const lower = p.toLowerCase()
    if (lower.startsWith('ckb') || lower === 'ku-iq' || lower.startsWith('ku-arab')) return 'ckb'
    const base = lower.split('-')[0]
    if (LANGUAGES.some((l) => l.code === base)) return base
  }
  return FALLBACK
}

export function resolveLanguage(setting: string): string {
  return setting && LANGUAGES.some((l) => l.code === setting) ? setting : detectSystemLanguage()
}

/** Applies lang/dir on <html> (drives fonts, logical CSS properties and rtl: variants). */
export function applyDocumentLanguage(code: string) {
  const def = getLanguage(code)
  const html = document.documentElement
  html.lang = def.code
  html.dir = def.dir
}

i18n.use(initReactI18next).init({
  resources: Object.fromEntries(LANGUAGES.map((l) => [l.code, { translation: l.resources }])),
  lng: detectSystemLanguage(),
  fallbackLng: FALLBACK,
  interpolation: { escapeValue: false }, // React escapes already
  returnNull: false,
})

i18n.on('languageChanged', applyDocumentLanguage)
applyDocumentLanguage(i18n.language)

export default i18n
