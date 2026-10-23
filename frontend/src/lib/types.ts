// TypeScript mirrors of the Go structs (json tags). Kept by hand on purpose: the app talks to
// window.go.main.App directly, so it does not depend on Wails' generated bindings layout.

export type ThemeMode = 'system' | 'dark' | 'light'
export type Accent = 'indigo' | 'violet' | 'blue' | 'teal' | 'rose' | 'amber'
export type Container = 'mp4' | 'mkv' | 'webm'
export type AudioFormat = 'mp3' | 'm4a' | 'opus' | 'flac' | 'wav'
export type AudioBitrate = 'best' | '320' | '256' | '192' | '128' | '96'

export interface Settings {
  language: string
  theme: ThemeMode
  accent: Accent
  nativeDigits: boolean
  calendar: 'gregorian' | 'jalali'
  translucency: boolean
  downloadDir: string
  rememberLastFolder: boolean
  lastFolder: string
  platformDirs: Record<string, string>
  filenameTemplate: string
  maxConcurrent: number
  speedLimitKBps: number
  defaultContainer: Container
  defaultAudioFormat: AudioFormat
  defaultAudioBitrate: AudioBitrate
  embedThumbnail: boolean
  embedMetadata: boolean
  embedChapters: boolean
  proxyEnabled: boolean
  proxyType: 'http' | 'https' | 'socks5'
  proxyHost: string
  proxyPort: number
  proxyUser: string
  proxyPass: string
  userAgent: string
  retries: number
  timeoutSec: number
  cookiesMode: 'none' | 'browser' | 'file'
  cookiesBrowser: string
  cookiesFile: string
  ytDlpPath: string
  ffmpegPath: string
  autoUpdateEngine: boolean
  clipboardWatch: boolean
  openFolderOnDone: boolean
  notify: boolean
  playSound: boolean
}

export interface PlatformDef {
  id: string
  name: string
  color: string
  icon: string
  hosts: string[]
  audioOnly: boolean
  matched: boolean
}

export interface PlatformInfo {
  id: string
  name: string
  color: string
  icon: string
  valid: boolean
  known: boolean
  audioOnly: boolean
  matched: boolean
  url: string
}

export interface AppInfo {
  name: string
  version: string
  os: 'windows' | 'darwin' | 'linux' | string
  arch: string
  translucent: boolean
  dataDir: string
  logPath: string
  platforms: PlatformDef[]
  templateFields: string[]
  cookieBrowsers: string[]
}

export interface Format {
  id: string
  ext: string
  width: number
  height: number
  fps: number
  hdr: boolean
  dynamicRange: string
  vcodec: string
  acodec: string
  tbr: number
  vbr: number
  abr: number
  asr: number
  channels: number
  filesize: number
  sizeApprox: boolean
  hasVideo: boolean
  hasAudio: boolean
  note: string
  language: string
  protocol: string
  label: string
}

export interface ResolutionGroup {
  height: number
  label: string
  formats: Format[]
}

export interface SubtitleTrack {
  lang: string
  name: string
  auto: boolean
  exts: string[]
}

export interface Thumbnail {
  url: string
  width: number
  height: number
  id: string
}

export interface SpotifyTrack {
  id: string
  title: string
  artists: string[]
  album: string
  durationMs: number
  coverUrl: string
  trackNumber: number
  releaseDate: string
  explicit: boolean
}

export interface Entry {
  index: number
  id: string
  title: string
  url: string
  duration: number
  thumbnail: string
  uploader: string
  spotify?: SpotifyTrack
}

export interface MediaInfo {
  id: string
  url: string
  title: string
  description: string
  uploader: string
  duration: number
  uploadDate: string
  viewCount: number
  likeCount: number
  thumbnail: string
  isLive: boolean
  platform: PlatformInfo
  extractor: string
  isPlaylist: boolean
  entries: Entry[]
  videoGroups: ResolutionGroup[]
  audio: Format[]
  highlights: { best: string; recommended: string; smallest: string }
  maxHeight: number
  subtitles: SubtitleTrack[]
  thumbnails: Thumbnail[]
  matched: boolean
  spotify?: { kind: string; id: string; title: string; owner: string; coverUrl: string; releaseDate: string; tracks: SpotifyTrack[] }
}

export type DownloadKind = 'video' | 'audio' | 'subtitles' | 'thumbnail'

export interface DownloadRequest {
  url: string
  title: string
  uploader: string
  thumbnail: string
  platform: string
  duration: number
  uploadDate: string
  mediaId: string
  playlistTitle: string
  playlistIndex: number
  playlistItem: number
  kind: DownloadKind
  formatId: string
  hasAudio: boolean
  maxHeight: number
  height: number
  fps: number
  container: Container
  audioFormat: AudioFormat
  audioBitrate: AudioBitrate
  audioFormatId: string
  subLangs: string[]
  subMode: 'embed' | 'file'
  subFormat: 'srt' | 'vtt'
  autoSubs: boolean
  thumbnailUrl: string
  thumbnailFormat: 'jpg' | 'png' | 'original'
  embedThumbnail: boolean
  embedMetadata: boolean
  embedChapters: boolean
  outputDir: string
  filenameTemplate: string
  qualityLabel: string
  spotify?: SpotifyTrack
  album: string
}

export type JobStatus =
  | 'queued'
  | 'downloading'
  | 'merging'
  | 'converting'
  | 'paused'
  | 'retrying'
  | 'done'
  | 'failed'
  | 'canceled'

export interface Job {
  id: string
  request: DownloadRequest
  status: JobStatus
  progress: number
  downloaded: number
  total: number
  speed: number
  eta: number
  part: number
  parts: number
  filePath: string
  outputBase: string
  ext: string
  errorCode: string
  errorDetail: string
  attempts: number
  createdAt: number
  startedAt: number
  finishedAt: number
}

export interface ComponentStatus {
  name: string
  installed: boolean
  path: string
  version: string
  source: 'custom' | 'managed' | 'system' | ''
  latest: string
  updateAvailable: boolean
  optional: boolean
}

export interface EngineStatus {
  ytDlp: ComponentStatus
  ffmpeg: ComponentStatus
  deno: ComponentStatus
  busy: boolean
  ready: boolean
}

export interface EngineProgress {
  component: string
  stage: 'checking' | 'downloading' | 'verifying' | 'extracting' | 'done' | 'error'
  percent: number
  downloaded: number
  total: number
  errorCode?: string
  error?: string
}

export interface HistoryEntry {
  id: string
  url: string
  title: string
  uploader: string
  platform: string
  kind: DownloadKind
  quality: string
  thumbnail: string
  filePath: string
  size: number
  status: 'done' | 'failed' | 'canceled'
  errorCode: string
  createdAt: number
  finishedAt: number
}

export interface HistoryQuery {
  search: string
  platform: string
  status: string
  limit: number
  offset: number
}

export interface HistoryPage {
  items: HistoryEntry[]
  total: number
}

export interface AppError {
  code: string
  detail?: string
}
