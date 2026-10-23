import { create } from 'zustand'

export type Route = 'home' | 'downloads' | 'history' | 'settings'
export type SettingsSection = 'appearance' | 'downloads' | 'network' | 'cookies' | 'engine' | 'behavior' | 'advanced'

export interface Toast {
  id: number
  kind: 'success' | 'error' | 'info'
  title: string
  body?: string
  action?: { label: string; run: () => void }
  duration?: number
}

interface UIState {
  route: Route
  settingsSection: SettingsSection | null
  toasts: Toast[]
  navigate(route: Route, section?: SettingsSection): void
  toast(t: Omit<Toast, 'id'>): void
  dismiss(id: number): void
}

let nextId = 1

export const useUI = create<UIState>((set, get) => ({
  route: 'home',
  settingsSection: null,
  toasts: [],
  navigate(route, section) {
    set({ route, settingsSection: section ?? null })
  },
  toast(t) {
    const id = nextId++
    set({ toasts: [...get().toasts.slice(-3), { ...t, id }] })
    setTimeout(() => get().dismiss(id), t.duration ?? (t.kind === 'error' ? 7000 : 4000))
  },
  dismiss(id) {
    set({ toasts: get().toasts.filter((x) => x.id !== id) })
  },
}))
