import { create } from 'zustand'
import { api, toAppError } from '@/lib/api'
import type { AppInfo, Settings } from '@/lib/types'
import i18n, { resolveLanguage } from '@/i18n'

interface SettingsState {
  settings: Settings | null
  app: AppInfo | null
  load(): Promise<void>
  /** Optimistically applies a patch, then persists it (debounced for text inputs). */
  update(patch: Partial<Settings>, opts?: { debounce?: boolean }): void
  reset(): Promise<void>
}

let saveTimer: ReturnType<typeof setTimeout> | null = null

async function persist(get: () => SettingsState, set: (p: Partial<SettingsState>) => void) {
  const s = get().settings
  if (!s) return
  try {
    const saved = await api.SaveSettings(s)
    // Keep what the user is typing; only adopt server-side normalization for numbers/enums.
    set({ settings: { ...saved, ...pickStrings(get().settings!, saved) } })
  } catch (e) {
    console.error('save settings failed', toAppError(e))
  }
}

function pickStrings(local: Settings, saved: Settings): Partial<Settings> {
  const out: Record<string, unknown> = {}
  for (const k of ['proxyHost', 'proxyUser', 'proxyPass', 'userAgent', 'filenameTemplate', 'ytDlpPath', 'ffmpegPath', 'cookiesFile'] as const) {
    if (local[k] !== saved[k] && (local[k] as string).trim() === (saved[k] as string)) out[k] = local[k]
  }
  return out as Partial<Settings>
}

export const useSettings = create<SettingsState>((set, get) => ({
  settings: null,
  app: null,
  async load() {
    const [settings, app] = await Promise.all([api.GetSettings(), api.GetAppInfo()])
    set({ settings, app })
    applyAppearance(settings, app)
    await i18n.changeLanguage(resolveLanguage(settings.language))
  },
  update(patch, opts) {
    const cur = get().settings
    if (!cur) return
    const next = { ...cur, ...patch }
    set({ settings: next })
    applyAppearance(next, get().app)
    if ('language' in patch) void i18n.changeLanguage(resolveLanguage(next.language))
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => void persist(get, set), opts?.debounce ? 450 : 0)
  },
  async reset() {
    const settings = await api.ResetSettings()
    set({ settings })
    applyAppearance(settings, get().app)
    await i18n.changeLanguage(resolveLanguage(settings.language))
  },
}))

const media = typeof window !== 'undefined' ? window.matchMedia('(prefers-color-scheme: dark)') : null

export function applyAppearance(s: Settings, app: AppInfo | null) {
  const html = document.documentElement
  const dark = s.theme === 'dark' || (s.theme === 'system' && (media?.matches ?? true))
  html.classList.toggle('dark', dark)
  html.dataset.accent = s.accent
  html.classList.toggle('translucent', !!app?.translucent)
  html.dataset.os = app?.os ?? ''
}

media?.addEventListener('change', () => {
  const { settings, app } = useSettings.getState()
  if (settings) applyAppearance(settings, app)
})
