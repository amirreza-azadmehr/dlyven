// Typed bridge to the Go backend (window.go.main.App) and the Wails runtime (window.runtime).
// When the page runs in a plain browser (`npm run dev` without Wails) a mock backend is used,
// so the UI can be designed and tested without the desktop shell.
import type {
  AppError,
  AppInfo,
  DownloadRequest,
  EngineStatus,
  HistoryPage,
  HistoryQuery,
  Job,
  MediaInfo,
  PlatformInfo,
  Settings,
} from './types'
import { createMockBackend } from './mock'

export interface Backend {
  GetAppInfo(): Promise<AppInfo>
  GetSettings(): Promise<Settings>
  SaveSettings(s: Settings): Promise<Settings>
  ResetSettings(): Promise<Settings>
  DetectPlatform(url: string): Promise<PlatformInfo>
  ExtractURLs(text: string): Promise<string[]>
  Analyze(url: string): Promise<MediaInfo>
  AnalyzeForBatch(url: string): Promise<MediaInfo>
  CancelAnalyze(): Promise<void>
  Enqueue(req: DownloadRequest): Promise<Job>
  EnqueueMany(reqs: DownloadRequest[]): Promise<Job[]>
  ListDownloads(): Promise<Job[]>
  PauseDownload(id: string): Promise<void>
  ResumeDownload(id: string): Promise<void>
  CancelDownload(id: string): Promise<void>
  RetryDownload(id: string): Promise<void>
  RemoveDownload(id: string): Promise<void>
  ClearFinishedDownloads(): Promise<void>
  PreviewFilename(tpl: string, vars: Record<string, string>, ext: string): Promise<string>
  DefaultFolderFor(platform: string): Promise<string>
  FileExists(path: string): Promise<boolean>
  OpenFile(path: string): Promise<void>
  ShowInFolder(path: string): Promise<void>
  OpenFolder(path: string): Promise<void>
  OpenLogsFolder(): Promise<void>
  ChooseFolder(current: string, title: string): Promise<string>
  ChooseCookiesFile(title: string): Promise<string>
  ChooseExecutable(title: string): Promise<string>
  GetEngineStatus(): Promise<EngineStatus>
  InstallEngine(): Promise<EngineStatus>
  UpdateEngine(force: boolean): Promise<EngineStatus>
  CheckEngineUpdates(): Promise<EngineStatus>
  ListHistory(q: HistoryQuery): Promise<HistoryPage>
  HistoryPlatforms(): Promise<string[]>
  DeleteHistory(id: string): Promise<void>
  ClearHistory(): Promise<void>
  GetDiagnostics(): Promise<string>
}

export interface Runtime {
  EventsOn(name: string, cb: (...data: any[]) => void): () => void
  WindowMinimise(): void
  WindowToggleMaximise(): void
  WindowIsMaximised(): Promise<boolean>
  Quit(): void
  ClipboardGetText(): Promise<string>
  ClipboardSetText(text: string): Promise<boolean>
  BrowserOpenURL(url: string): void
}

declare global {
  interface Window {
    go?: { main?: { App?: Backend } }
    runtime?: Runtime
  }
}

export const isWails = typeof window !== 'undefined' && !!window.go?.main?.App

const mock = isWails ? null : createMockBackend()

export const api: Backend = isWails ? window.go!.main!.App! : mock!.backend

export const rt: Runtime = isWails
  ? window.runtime!
  : {
      EventsOn: (name, cb) => mock!.on(name, cb),
      WindowMinimise: () => {},
      WindowToggleMaximise: () => {},
      WindowIsMaximised: async () => false,
      Quit: () => {},
      ClipboardGetText: async () => {
        try {
          return await navigator.clipboard.readText()
        } catch {
          return ''
        }
      },
      ClipboardSetText: async (text) => {
        try {
          await navigator.clipboard.writeText(text)
          return true
        } catch {
          return false
        }
      },
      BrowserOpenURL: (url) => window.open(url, '_blank', 'noopener'),
    }

/** Parses a rejected backend promise into {code, detail}. Go errors serialize as JSON. */
export function toAppError(err: unknown): AppError {
  if (err && typeof err === 'object' && 'code' in err) return err as AppError
  const s = typeof err === 'string' ? err : err instanceof Error ? err.message : String(err)
  try {
    const parsed = JSON.parse(s)
    if (parsed && typeof parsed.code === 'string') return parsed
  } catch {
    /* not JSON */
  }
  return { code: 'ERR_UNKNOWN', detail: s }
}
