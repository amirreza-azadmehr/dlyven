import { create } from 'zustand'
import { api, rt } from '@/lib/api'
import type { Job } from '@/lib/types'

interface DownloadsState {
  jobs: Record<string, Job>
  order: string[]
  init(): Promise<() => void>
}

export const ACTIVE: Job['status'][] = ['downloading', 'merging', 'converting', 'retrying']
export const isActive = (j: Job) => ACTIVE.includes(j.status)
export const isFinal = (j: Job) => j.status === 'done' || j.status === 'failed' || j.status === 'canceled'

export const useDownloads = create<DownloadsState>((set, get) => ({
  jobs: {},
  order: [],
  async init() {
    const list = await api.ListDownloads()
    set({ jobs: Object.fromEntries(list.map((j) => [j.id, j])), order: list.map((j) => j.id) })
    const offUpdate = rt.EventsOn('download:update', (j: Job) => {
      const { jobs, order } = get()
      set({ jobs: { ...jobs, [j.id]: j }, order: order.includes(j.id) ? order : [...order, j.id] })
    })
    const offRemoved = rt.EventsOn('download:removed', (id: string) => {
      const { jobs, order } = get()
      const next = { ...jobs }
      delete next[id]
      set({ jobs: next, order: order.filter((x) => x !== id) })
    })
    return () => {
      offUpdate()
      offRemoved()
    }
  },
}))

/** Number of jobs that are not finished yet (shown as the sidebar badge). */
export const useActiveCount = () =>
  useDownloads((s) => s.order.reduce((n, id) => (s.jobs[id] && !isFinal(s.jobs[id]) ? n + 1 : n), 0))
