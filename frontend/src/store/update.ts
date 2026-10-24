import { create } from 'zustand'
import { api, rt, toAppError } from '@/lib/api'
import type { AppError, UpdateProgress, UpdateStatus } from '@/lib/types'

interface UpdateState {
  status: UpdateStatus | null
  progress: UpdateProgress | null
  error: AppError | null
  checking: boolean
  downloading: boolean
  init(): Promise<() => void>
  check(): Promise<UpdateStatus | null>
  download(): Promise<boolean>
  restart(): Promise<void>
}

/** Dlyven's own updates (the engine has its own store). The backend checks on launch. */
export const useUpdate = create<UpdateState>((set) => ({
  status: null,
  progress: null,
  error: null,
  checking: false,
  downloading: false,
  async init() {
    set({ status: await api.GetUpdateStatus() })
    const offStatus = rt.EventsOn('update:status', (status: UpdateStatus) => set({ status }))
    const offProgress = rt.EventsOn('update:progress', (p: UpdateProgress) => {
      set({ progress: p })
      if (p.stage === 'error') set({ error: { code: p.errorCode || 'ERR_UNKNOWN', detail: p.error } })
    })
    return () => {
      offStatus()
      offProgress()
    }
  },
  async check() {
    set({ checking: true, error: null })
    try {
      const status = await api.CheckForUpdates()
      set({ status })
      return status
    } catch (e) {
      set({ error: toAppError(e) })
      return null
    } finally {
      set({ checking: false })
    }
  },
  async download() {
    set({ downloading: true, error: null, progress: null })
    try {
      set({ status: await api.DownloadUpdate() })
      return true
    } catch (e) {
      set({ error: toAppError(e) })
      return false
    } finally {
      set({ downloading: false })
    }
  },
  async restart() {
    try {
      await api.RestartToUpdate()
    } catch (e) {
      set({ error: toAppError(e) })
    }
  },
}))
