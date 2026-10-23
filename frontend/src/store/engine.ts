import { create } from 'zustand'
import { api, rt, toAppError } from '@/lib/api'
import type { AppError, EngineProgress, EngineStatus } from '@/lib/types'

interface EngineState {
  status: EngineStatus | null
  progress: Record<string, EngineProgress>
  error: AppError | null
  working: boolean
  init(): Promise<() => void>
  refresh(): Promise<void>
  install(): Promise<void>
  update(force?: boolean): Promise<boolean>
  check(): Promise<EngineStatus | null>
}

export const useEngine = create<EngineState>((set, get) => ({
  status: null,
  progress: {},
  error: null,
  working: false,
  async init() {
    await get().refresh()
    const offStatus = rt.EventsOn('engine:status', (status: EngineStatus) => set({ status }))
    const offProgress = rt.EventsOn('engine:progress', (p: EngineProgress) => {
      set({ progress: { ...get().progress, [p.component]: p } })
      if (p.stage === 'error') set({ error: { code: p.errorCode || 'ERR_ENGINE_INSTALL', detail: p.error } })
      if (p.stage === 'done') void get().refresh()
    })
    return () => {
      offStatus()
      offProgress()
    }
  },
  async refresh() {
    set({ status: await api.GetEngineStatus() })
  },
  async install() {
    set({ working: true, error: null, progress: {} })
    try {
      set({ status: await api.InstallEngine() })
    } catch (e) {
      set({ error: toAppError(e) })
    } finally {
      set({ working: false })
    }
  },
  async update(force = false) {
    set({ working: true, error: null, progress: {} })
    try {
      set({ status: await api.UpdateEngine(force) })
      return true
    } catch (e) {
      set({ error: toAppError(e) })
      return false
    } finally {
      set({ working: false })
    }
  },
  async check() {
    try {
      const status = await api.CheckEngineUpdates()
      set({ status })
      return status
    } catch (e) {
      set({ error: toAppError(e) })
      return null
    }
  },
}))
