import { create } from 'zustand'
import { api, toAppError } from '@/lib/api'
import type { AppError, AudioBitrate, AudioFormat, Container, MediaInfo } from '@/lib/types'
import { useSettings } from './settings'

export type Tab = 'video' | 'audio' | 'subtitles' | 'thumbnail'
export type PresetId = 'best' | '2160' | '1080' | '720' | '480' | 'mp3' | 'm4a'

export interface Selection {
  tab: Tab
  /** Exact video format (single items). null → use maxHeight preset. */
  formatId: string | null
  /** 0 = best available. */
  maxHeight: number
  preset: PresetId | null
  container: Container
  audioFormat: AudioFormat
  audioBitrate: AudioBitrate
  audioFormatId: string
  /** Subtitle keys: `${lang}|${auto ? 'a' : 'm'}` */
  subs: string[]
  subMode: 'embed' | 'file'
  subFormat: 'srt' | 'vtt'
  thumbUrl: string
  thumbFormat: 'jpg' | 'png' | 'original'
  embedThumbnail: boolean
  embedMetadata: boolean
  /** Selected playlist entry indexes (1-based). */
  entries: number[]
  outputDir: string
}

type Status = 'idle' | 'analyzing' | 'ready' | 'error'

interface AnalysisState {
  url: string
  status: Status
  info: MediaInfo | null
  error: AppError | null
  sel: Selection
  setUrl(url: string): void
  analyze(url?: string): Promise<void>
  cancel(): void
  reset(): void
  select(patch: Partial<Selection>): void
  applyPreset(p: PresetId): void
}

function initialSelection(): Selection {
  const s = useSettings.getState().settings
  return {
    tab: 'video',
    formatId: null,
    maxHeight: 0,
    preset: 'best',
    container: s?.defaultContainer ?? 'mp4',
    audioFormat: s?.defaultAudioFormat ?? 'mp3',
    audioBitrate: s?.defaultAudioBitrate ?? '320',
    audioFormatId: '',
    subs: [],
    subMode: 'embed',
    subFormat: 'srt',
    thumbUrl: '',
    thumbFormat: 'jpg',
    embedThumbnail: s?.embedThumbnail ?? true,
    embedMetadata: s?.embedMetadata ?? true,
    entries: [],
    outputDir: '',
  }
}

let analyzeSeq = 0

export const useAnalysis = create<AnalysisState>((set, get) => ({
  url: '',
  status: 'idle',
  info: null,
  error: null,
  sel: initialSelection(),
  setUrl(url) {
    set({ url })
  },
  async analyze(urlArg) {
    const url = (urlArg ?? get().url).trim()
    if (!url) return
    const seq = ++analyzeSeq
    set({ url, status: 'analyzing', error: null, info: null })
    try {
      const info = await api.Analyze(url)
      if (seq !== analyzeSeq) return
      const sel = initialSelection()
      if (info.platform.audioOnly || info.matched || info.videoGroups.length === 0) {
        sel.tab = 'audio'
        sel.preset = sel.audioFormat === 'm4a' ? 'm4a' : sel.audioFormat === 'mp3' ? 'mp3' : null
      }
      if (!info.isPlaylist && info.highlights.recommended && info.videoGroups.length) {
        sel.formatId = info.highlights.recommended
        sel.preset = null
      }
      sel.entries = info.entries.map((e) => e.index)
      sel.thumbUrl = info.thumbnails[0]?.url ?? info.thumbnail
      sel.outputDir = await api.DefaultFolderFor(info.platform.id)
      set({ status: 'ready', info, sel })
    } catch (e) {
      if (seq !== analyzeSeq) return
      const err = toAppError(e)
      if (err.code === 'ERR_CANCELED') set({ status: 'idle' })
      else set({ status: 'error', error: err })
    }
  },
  cancel() {
    analyzeSeq++
    void api.CancelAnalyze()
    set({ status: 'idle', error: null })
  },
  reset() {
    analyzeSeq++
    if (get().status === 'analyzing') void api.CancelAnalyze()
    set({ url: '', status: 'idle', info: null, error: null, sel: initialSelection() })
  },
  select(patch) {
    set({ sel: { ...get().sel, ...patch } })
  },
  applyPreset(p) {
    const { sel } = get()
    switch (p) {
      case 'mp3':
      case 'm4a':
        set({ sel: { ...sel, preset: p, tab: 'audio', audioFormat: p, audioFormatId: '', audioBitrate: p === 'mp3' ? '320' : 'best' } })
        return
      case 'best':
        set({ sel: { ...sel, preset: p, tab: 'video', formatId: null, maxHeight: 0 } })
        return
      default:
        set({ sel: { ...sel, preset: p, tab: 'video', formatId: null, maxHeight: Number(p) } })
    }
  },
}))
