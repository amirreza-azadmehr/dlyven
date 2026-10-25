// In-browser mock of the Go backend, used only when running outside Wails (`npm run dev`).
// It simulates analysis results, a download queue with progress, history and settings.
import type { Backend } from './api'
import type {
  AppInfo,
  DownloadRequest,
  EngineStatus,
  Format,
  HistoryEntry,
  Job,
  MediaInfo,
  PlatformDef,
  PlatformInfo,
  Settings,
  UpdateStatus,
} from './types'

const platforms: PlatformDef[] = [
  { id: 'youtube', name: 'YouTube', color: '#FF4E45', icon: 'square-play', hosts: ['youtube.com', 'youtu.be', 'music.youtube.com'], audioOnly: false, matched: false },
  { id: 'instagram', name: 'Instagram', color: '#E1306C', icon: 'camera', hosts: ['instagram.com', 'instagr.am', 'ddinstagram.com'], audioOnly: false, matched: false },
  { id: 'spotify', name: 'Spotify', color: '#1DB954', icon: 'music', hosts: ['open.spotify.com', 'spotify.link'], audioOnly: true, matched: true },
  { id: 'tiktok', name: 'TikTok', color: '#25F4EE', icon: 'music-2', hosts: ['tiktok.com'], audioOnly: false, matched: false },
  { id: 'twitter', name: 'X / Twitter', color: '#8899A6', icon: 'at-sign', hosts: ['twitter.com', 'x.com'], audioOnly: false, matched: false },
  { id: 'facebook', name: 'Facebook', color: '#1877F2', icon: 'users', hosts: ['facebook.com', 'fb.watch'], audioOnly: false, matched: false },
  { id: 'vimeo', name: 'Vimeo', color: '#1AB7EA', icon: 'video', hosts: ['vimeo.com'], audioOnly: false, matched: false },
  { id: 'soundcloud', name: 'SoundCloud', color: '#FF5500', icon: 'cloud', hosts: ['soundcloud.com'], audioOnly: true, matched: false },
  { id: 'reddit', name: 'Reddit', color: '#FF4500', icon: 'message-circle', hosts: ['reddit.com', 'redd.it'], audioOnly: false, matched: false },
  { id: 'twitch', name: 'Twitch', color: '#9146FF', icon: 'gamepad-2', hosts: ['twitch.tv'], audioOnly: false, matched: false },
]

function detect(raw: string): PlatformInfo {
  let s = raw.trim()
  if (!s) return { id: '', name: '', color: '', icon: '', valid: false, known: false, audioOnly: false, matched: false, url: '' }
  if (!/^[a-z]+:\/\//i.test(s) && /^([a-z0-9-]+\.)+[a-z]{2,}(\/|$)/i.test(s)) s = 'https://' + s
  let host = ''
  try {
    const u = new URL(s)
    if (!/^https?:$/.test(u.protocol) || !u.hostname.includes('.')) throw new Error()
    host = u.hostname.replace(/^www\./, '')
  } catch {
    return { id: '', name: '', color: '', icon: '', valid: false, known: false, audioOnly: false, matched: false, url: '' }
  }
  const d = platforms.find((p) => p.hosts.some((h) => host === h || host.endsWith('.' + h)))
  if (d) return { ...d, valid: true, known: true, url: s }
  return { id: 'generic', name: host, color: '#8B93A7', icon: 'globe', valid: true, known: false, audioOnly: false, matched: false, url: s }
}

const defaults: Settings = {
  language: '', theme: 'system', accent: 'indigo', nativeDigits: false, calendar: 'gregorian', translucency: true,
  downloadDir: '/home/user/Downloads', rememberLastFolder: true, lastFolder: '', platformDirs: {},
  filenameTemplate: '{title} - {quality}', maxConcurrent: 3, speedLimitKBps: 0, defaultContainer: 'mp4',
  defaultAudioFormat: 'mp3', defaultAudioBitrate: '320', embedThumbnail: true, embedMetadata: true, embedChapters: true,
  proxyEnabled: false, proxyType: 'http', proxyHost: '', proxyPort: 0, proxyUser: '', proxyPass: '', userAgent: '',
  retries: 3, timeoutSec: 30, cookiesMode: 'none', cookiesBrowser: 'chrome', cookiesFile: '', ytDlpPath: '', ffmpegPath: '',
  autoUpdateEngine: true, autoCheckUpdates: true, clipboardWatch: false, openFolderOnDone: false, notify: true, playSound: true,
}

function vf(id: string, h: number, fps: number, vcodec: string, tbr: number, ext = 'mp4', hdr = false, audio = false): Format {
  const w = Math.round((h * 16) / 9)
  return {
    id, ext, width: w, height: h, fps, hdr, dynamicRange: hdr ? 'HDR10' : 'SDR', vcodec, acodec: audio ? 'AAC' : '', tbr, vbr: tbr, abr: 0, asr: 0, channels: 0,
    filesize: Math.round((tbr * 1000 * 212) / 8), sizeApprox: id.endsWith('a'), hasVideo: true, hasAudio: audio, note: '', language: '', protocol: 'https',
    label: `${h}p${fps > 30 ? fps : ''}`,
  }
}
function af(id: string, ext: string, codec: string, abr: number): Format {
  return { id, ext, width: 0, height: 0, fps: 0, hdr: false, dynamicRange: '', vcodec: '', acodec: codec, tbr: abr, vbr: 0, abr, asr: 48000, channels: 2,
    filesize: Math.round((abr * 1000 * 212) / 8), sizeApprox: false, hasVideo: false, hasAudio: true, note: '', language: 'en', protocol: 'https', label: '' }
}

function sampleVideo(p: PlatformInfo): MediaInfo {
  const groups = [
    { height: 2160, label: '2160p', formats: [vf('337', 2160, 60, 'VP9', 30000, 'webm', true), vf('315', 2160, 60, 'VP9', 24000, 'webm'), vf('401', 2160, 60, 'AV1', 21000, 'mp4')] },
    { height: 1440, label: '1440p', formats: [vf('308', 1440, 60, 'VP9', 13000, 'webm'), vf('400', 1440, 60, 'AV1', 11000, 'mp4')] },
    { height: 1080, label: '1080p', formats: [vf('299', 1080, 60, 'H.264', 6000), vf('303', 1080, 60, 'VP9', 4500, 'webm'), vf('137', 1080, 30, 'H.264', 4200), vf('248a', 1080, 30, 'VP9', 2900, 'webm')] },
    { height: 720, label: '720p', formats: [vf('298', 720, 60, 'H.264', 3400), vf('22', 720, 30, 'H.264', 1800, 'mp4', false, true), vf('247', 720, 30, 'VP9', 1500, 'webm')] },
    { height: 480, label: '480p', formats: [vf('135', 480, 30, 'H.264', 1100), vf('244', 480, 30, 'VP9', 750, 'webm')] },
    { height: 360, label: '360p', formats: [vf('18', 360, 30, 'H.264', 600, 'mp4', false, true), vf('243', 360, 30, 'VP9', 420, 'webm')] },
    { height: 144, label: '144p', formats: [vf('160', 144, 30, 'H.264', 110)] },
  ]
  return {
    id: 'demo', url: p.url, title: 'Northern Lights over Lofoten — 4K Timelapse', description: '', uploader: 'Aurora Films',
    duration: 212, uploadDate: '2024-03-14', viewCount: 1843211, likeCount: 52000,
    thumbnail: thumb('#4f46e5', '#0ea5e9'), isLive: false, platform: p, extractor: 'Youtube', isPlaylist: false, entries: [],
    videoGroups: groups, audio: [af('251', 'webm', 'Opus', 160), af('140', 'm4a', 'AAC', 129), af('250', 'webm', 'Opus', 70), af('139', 'm4a', 'AAC', 48)],
    highlights: { best: '337', recommended: '299', smallest: '160' }, maxHeight: 2160,
    subtitles: [
      { lang: 'en', name: 'English', auto: false, exts: ['vtt', 'srt'] },
      { lang: 'fa', name: 'فارسی', auto: false, exts: ['vtt'] },
      { lang: 'tr', name: 'Türkçe', auto: false, exts: ['vtt'] },
      { lang: 'en', name: 'English (auto-generated)', auto: true, exts: ['vtt'] },
      { lang: 'de', name: 'German', auto: true, exts: ['vtt'] },
    ],
    thumbnails: [
      { url: thumb('#4f46e5', '#0ea5e9'), width: 1280, height: 720, id: 'maxres' },
      { url: thumb('#4f46e5', '#0ea5e9'), width: 640, height: 480, id: 'sd' },
      { url: thumb('#4f46e5', '#0ea5e9'), width: 320, height: 180, id: 'mq' },
    ],
    matched: false,
  }
}

function thumb(a: string, b: string): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${a}"/><stop offset="1" stop-color="${b}"/></linearGradient></defs><rect width="640" height="360" fill="url(#g)"/><path d="M0 270 Q160 200 320 250 T640 230 V360 H0Z" fill="#000" opacity=".35"/></svg>`
  return 'data:image/svg+xml;utf8,' + encodeURIComponent(svg)
}

function samplePlaylist(p: PlatformInfo): MediaInfo {
  const titles = ['Intro — Field Recordings', 'Night Drive', 'Journey North', 'Northern Lights', 'New Day', 'Afterglow', 'Low Tide', 'Outro']
  return {
    ...sampleVideo(p), title: 'Late Night Mix 2024', isPlaylist: true, videoGroups: [], audio: [], subtitles: [], maxHeight: 0,
    entries: titles.map((t, i) => ({ index: i + 1, id: 'v' + i, title: t, url: p.url + '&i=' + i, duration: 150 + i * 23, thumbnail: thumb(['#f43f5e', '#8b5cf6', '#06b6d4', '#22c55e'][i % 4], '#111827'), uploader: 'Various' })),
  }
}

function sampleSpotify(p: PlatformInfo): MediaInfo {
  const artists = ['Nova Atlas']
  const tracks = ['Glass Harbor', 'Paper Moons', 'Coastline'].map((t, i) => ({
    id: 's' + i, title: t, artists, album: 'Tidal Lines', durationMs: 180000 + i * 15000, coverUrl: thumb('#16a34a', '#0f172a'), trackNumber: i + 1, releaseDate: '2023-09-01', explicit: false,
  }))
  return {
    ...sampleVideo(p), title: 'Tidal Lines', uploader: 'Nova Atlas', isPlaylist: true, matched: true, videoGroups: [], audio: [], subtitles: [],
    thumbnail: thumb('#16a34a', '#0f172a'), thumbnails: [{ url: thumb('#16a34a', '#0f172a'), width: 640, height: 640, id: 'cover' }],
    entries: tracks.map((t, i) => ({ index: i + 1, id: t.id, title: t.title, url: '', duration: t.durationMs / 1000, thumbnail: t.coverUrl, uploader: 'Nova Atlas', spotify: t })),
    spotify: { kind: 'album', id: 'x', title: 'Tidal Lines', owner: 'Nova Atlas', coverUrl: thumb('#16a34a', '#0f172a'), releaseDate: '2023-09-01', tracks },
  }
}

/** Fills the queue and history with representative items (?seed=1). */
function seedDemo(jobs: Map<string, Job>, history: HistoryEntry[]) {
  const base = (title: string, platform: string, kind: DownloadRequest['kind'], quality: string, color: string): DownloadRequest => ({
    url: 'https://example.com/' + encodeURIComponent(title), title, uploader: 'Demo', thumbnail: thumb(color, '#0f172a'), platform, duration: 200,
    uploadDate: '2024-05-01', mediaId: title, playlistTitle: '', playlistIndex: 0, playlistItem: 0, kind, formatId: '', hasAudio: false,
    maxHeight: 0, height: 0, fps: 0, container: 'mp4', audioFormat: 'mp3', audioBitrate: '320', audioFormatId: '', subLangs: [], subMode: 'embed',
    subFormat: 'srt', autoSubs: false, thumbnailUrl: '', thumbnailFormat: 'jpg', embedThumbnail: true, embedMetadata: true, embedChapters: true,
    outputDir: '/home/user/Downloads', filenameTemplate: '{title} - {quality}', qualityLabel: quality, album: '',
  })
  const mk = (id: string, req: DownloadRequest, patch: Partial<Job>): Job => ({
    id, request: req, status: 'queued', progress: 0, downloaded: 0, total: 0, speed: 0, eta: -1, part: 1, parts: 1, filePath: '',
    outputBase: req.title, ext: 'mp4', errorCode: '', errorDetail: '', attempts: 1, createdAt: Date.now() - 60000, startedAt: 0, finishedAt: 0, ...patch,
  })
  const demo: Job[] = [
    mk('d1', base('Northern Lights over Lofoten — 4K Timelapse', 'youtube', 'video', '2160p60 HDR', '#4f46e5'), { status: 'paused', progress: 64, downloaded: 412_000_000, total: 640_000_000, part: 1, parts: 2 }),
    mk('d2', base('Journey North — Full Documentary', 'youtube', 'video', '1080p', '#06b6d4'), { status: 'merging', progress: 99.9, downloaded: 220_000_000, total: 221_000_000 }),
    mk('d3', base('Live Performance (Clip)', 'instagram', 'video', '720p', '#e11d48'), { status: 'failed', errorCode: 'ERR_LOGIN_REQUIRED' }),
    mk('d4', base('Glass Harbor', 'spotify', 'audio', 'MP3 320k', '#16a34a'), { status: 'done', progress: 100, total: 8_400_000, finishedAt: Date.now() - 3600_000, filePath: '/home/user/Music/Glass Harbor.mp3' }),
    mk('d5', base('New Day — Acoustic', 'soundcloud', 'audio', 'M4A', '#f97316'), { status: 'queued' }),
  ]
  demo[3].request.spotify = { id: 's', title: 'Glass Harbor', artists: ['Nova Atlas'], album: 'Tidal Lines', durationMs: 180000, coverUrl: '', trackNumber: 1, releaseDate: '2023', explicit: false }
  for (const j of demo) jobs.set(j.id, j)
  for (const j of demo.filter((x) => x.status === 'done' || x.status === 'failed')) {
    history.push({ id: j.id, url: j.request.url, title: j.request.title, uploader: 'Demo', platform: j.request.platform, kind: j.request.kind, quality: j.request.qualityLabel,
      thumbnail: j.request.thumbnail, filePath: j.filePath, size: j.total, status: j.status as 'done' | 'failed', errorCode: j.errorCode, createdAt: j.createdAt, finishedAt: Date.now() - 7200_000 })
  }
}

export function createMockBackend() {
  const listeners = new Map<string, Set<(...a: any[]) => void>>()
  const emit = (name: string, data: unknown) => listeners.get(name)?.forEach((cb) => cb(data))
  let settings: Settings = { ...defaults }
  try {
    const saved = localStorage.getItem('dlyven-mock-settings')
    if (saved) settings = { ...defaults, ...JSON.parse(saved) }
  } catch {
    /* ignore */
  }
  // Dev preview: ?lang=fa&theme=light&accent=rose lets designers open any variant directly.
  const q = new URLSearchParams(location.search)
  if (q.get('lang') !== null) settings.language = q.get('lang')!
  if (q.get('theme')) settings.theme = q.get('theme') as Settings['theme']
  if (q.get('accent')) settings.accent = q.get('accent') as Settings['accent']
  if (q.get('digits')) settings.nativeDigits = q.get('digits') === '1'
  const jobs = new Map<string, Job>()
  const history: HistoryEntry[] = []
  let seq = 0
  if (q.get('seed')) seedDemo(jobs, history)
  let analyzeTimer: ReturnType<typeof setTimeout> | null = null
  let analyzeReject: ((e: unknown) => void) | null = null

  const engine: EngineStatus = {
    ytDlp: { name: 'yt-dlp', installed: true, path: '~/Dlyven/bin/yt-dlp', version: '2026.10.21', source: 'managed', latest: '2026.10.21', updateAvailable: false, optional: false },
    ffmpeg: { name: 'ffmpeg', installed: true, path: '~/Dlyven/bin/ffmpeg', version: 'N-127941', source: 'managed', latest: '', updateAvailable: false, optional: false },
    deno: { name: 'deno', installed: true, path: '~/Dlyven/bin/deno', version: '2.9.8', source: 'managed', latest: '', updateAvailable: false, optional: true },
    busy: false, ready: true,
  }

  // ?update=1 pretends a newer release exists; ?update=manual shows the download-it-yourself
  // flow; ?update=ready opens with the update downloaded and waiting for a restart.
  const update: UpdateStatus = {
    current: '1.0.2', latest: '', available: false, mode: q.get('update') === 'manual' ? 'manual' : 'replace',
    asset: 'Dlyven-macOS-Universal.zip', releaseUrl: 'https://github.com/amirreza-azadmehr/dlyven/releases/latest', ready: false, busy: false,
  }
  const checkUpdate = () => {
    if (q.get('update')) Object.assign(update, { latest: '1.0.3', available: true, ready: q.get('update') === 'ready', releaseUrl: 'https://github.com/amirreza-azadmehr/dlyven/releases/tag/v1.0.3' })
    else update.latest = update.current
    emit('update:status', { ...update })
    return { ...update }
  }
  if (q.get('update')) setTimeout(checkUpdate, 1200)

  function tick(id: string) {
    const j = jobs.get(id)
    if (!j || j.status !== 'downloading') return
    j.progress = Math.min(100, j.progress + 2 + Math.random() * 5)
    j.total = 84_000_000
    j.downloaded = Math.round((j.total * j.progress) / 100)
    j.speed = 3_500_000 + Math.random() * 2_000_000
    j.eta = Math.round((j.total - j.downloaded) / j.speed)
    j.part = j.progress < 70 ? 1 : 2
    j.parts = j.request.kind === 'video' && !j.request.hasAudio ? 2 : 1
    if (j.progress >= 100) {
      j.status = 'merging'
      emit('download:update', { ...j })
      setTimeout(() => {
        j.status = 'done'
        j.progress = 100
        j.filePath = `${j.request.outputDir}/${j.outputBase}.${j.ext}`
        j.finishedAt = Date.now()
        emit('download:update', { ...j })
        emit('download:finished', { ...j })
        history.unshift({ id: j.id, url: j.request.url, title: j.request.title, uploader: j.request.uploader, platform: j.request.platform, kind: j.request.kind,
          quality: j.request.qualityLabel, thumbnail: j.request.thumbnail, filePath: j.filePath, size: j.total, status: 'done', errorCode: '', createdAt: j.createdAt, finishedAt: j.finishedAt })
      }, 900)
      return
    }
    emit('download:update', { ...j })
    setTimeout(() => tick(id), 250)
  }

  function schedule() {
    const running = [...jobs.values()].filter((j) => j.status === 'downloading').length
    let slots = settings.maxConcurrent - running
    for (const j of jobs.values()) {
      if (slots <= 0) break
      if (j.status === 'queued') {
        j.status = 'downloading'
        j.startedAt = Date.now()
        emit('download:update', { ...j })
        setTimeout(() => tick(j.id), 200)
        slots--
      }
    }
  }

  const backend: Backend = {
    async GetAppInfo(): Promise<AppInfo> {
      return { name: 'Dlyven', version: '1.0.0-dev', os: navigator.platform.toLowerCase().includes('mac') ? 'darwin' : 'windows', arch: 'amd64', translucent: false,
        dataDir: '~/.config/Dlyven', logPath: '~/.config/Dlyven/logs/dlyven.log', platforms,
        templateFields: ['title', 'uploader', 'quality', 'resolution', 'fps', 'platform', 'id', 'date', 'playlist', 'index', 'artist', 'album', 'track'],
        cookieBrowsers: ['chrome', 'edge', 'firefox', 'brave', 'chromium', 'opera', 'vivaldi', 'safari'],
        installedBrowsers: ['chrome', 'firefox'] }
    },
    async GetSettings() { return { ...settings } },
    async SaveSettings(s) {
      settings = { ...s, maxConcurrent: Math.min(10, Math.max(1, s.maxConcurrent)) }
      localStorage.setItem('dlyven-mock-settings', JSON.stringify(settings))
      schedule()
      return { ...settings }
    },
    async ResetSettings() { settings = { ...defaults, language: settings.language }; return { ...settings } },
    async DetectPlatform(url) { return detect(url) },
    async ExtractURLs(text) {
      const out: string[] = []
      for (const m of text.match(/https?:\/\/[^\s<>"']+/gi) ?? []) {
        const u = m.replace(/[.,;:!?)\]]+$/, '')
        if (!out.includes(u)) out.push(u)
      }
      return out
    },
    Analyze(url) {
      return new Promise<MediaInfo>((resolve, reject) => {
        const p = detect(url)
        analyzeReject = reject
        const delay = q.has('shot') ? 0 : 1400
        analyzeTimer = setTimeout(() => {
          analyzeReject = null
          if (!p.valid) return reject(JSON.stringify({ code: 'ERR_INVALID_URL' }))
          if (/error|private/.test(url)) return reject(JSON.stringify({ code: 'ERR_LOGIN_REQUIRED', detail: 'ERROR: [Instagram] Requested content is not available, rate-limit reached or login required' }))
          if (p.id === 'spotify') return resolve(sampleSpotify(p))
          if (/list=|playlist|album/.test(url)) return resolve(samplePlaylist(p))
          resolve(sampleVideo(p))
        }, delay)
      })
    },
    async AnalyzeForBatch(url) {
      await new Promise((r) => setTimeout(r, 500))
      const p = detect(url)
      if (!p.valid) throw JSON.stringify({ code: 'ERR_INVALID_URL' })
      return p.id === 'spotify' ? sampleSpotify(p) : { ...sampleVideo(p), title: decodeURIComponent(url.split('/').pop() || 'Video') }
    },
    async CancelAnalyze() {
      if (analyzeTimer) clearTimeout(analyzeTimer)
      analyzeReject?.(JSON.stringify({ code: 'ERR_CANCELED' }))
      analyzeReject = null
    },
    async Enqueue(req: DownloadRequest) {
      const id = 'job' + ++seq
      const ext = req.kind === 'audio' ? req.audioFormat : req.kind === 'subtitles' ? req.subFormat : req.kind === 'thumbnail' ? 'jpg' : req.container
      const j: Job = { id, request: { ...req, outputDir: req.outputDir || settings.downloadDir }, status: 'queued', progress: 0, downloaded: 0, total: 0, speed: 0, eta: -1, part: 1, parts: 1,
        filePath: '', outputBase: `${req.title} - ${req.qualityLabel}`, ext, errorCode: '', errorDetail: '', attempts: 0, createdAt: Date.now(), startedAt: 0, finishedAt: 0 }
      jobs.set(id, j)
      emit('download:update', { ...j })
      setTimeout(schedule, 50)
      return { ...j }
    },
    async EnqueueMany(reqs) { const out: Job[] = []; for (const r of reqs) out.push(await backend.Enqueue(r)); return out },
    async ListDownloads() { return [...jobs.values()].map((j) => ({ ...j })) },
    async PauseDownload(id) { const j = jobs.get(id); if (j && !['done', 'failed', 'canceled'].includes(j.status)) { j.status = 'paused'; j.speed = 0; emit('download:update', { ...j }); schedule() } },
    async ResumeDownload(id) { const j = jobs.get(id); if (j?.status === 'paused') { j.status = 'queued'; emit('download:update', { ...j }); schedule() } },
    async CancelDownload(id) { const j = jobs.get(id); if (j) { j.status = 'canceled'; j.speed = 0; emit('download:update', { ...j }); schedule() } },
    async RetryDownload(id) { const j = jobs.get(id); if (j) { j.status = 'queued'; j.progress = 0; j.errorCode = ''; emit('download:update', { ...j }); schedule() } },
    async RemoveDownload(id) { jobs.delete(id); emit('download:removed', id); schedule() },
    async ClearFinishedDownloads() { for (const [id, j] of jobs) if (['done', 'failed', 'canceled'].includes(j.status)) { jobs.delete(id); emit('download:removed', id) } },
    async PreviewFilename(tpl, vars, ext) {
      let s = tpl.replace(/\.\{ext\}$/, '').replace(/\{([a-z]+)\}/g, (_, k) => vars[k] ?? '')
      s = s.replace(/[\\/:*?"<>|]/g, '-').replace(/(\s*[-_|]\s*){2,}/g, ' - ').replace(/\s{2,}/g, ' ').trim().replace(/^[-_|\s]+|[-_|\s]+$/g, '')
      return (s || 'download') + (ext ? '.' + ext : '')
    },
    async DefaultFolderFor(p) { return settings.platformDirs[p] || (settings.rememberLastFolder && settings.lastFolder) || settings.downloadDir },
    async FileExists() { return true },
    async OpenFile() {}, async ShowInFolder() {}, async OpenFolder() {}, async OpenLogsFolder() {},
    async ChooseFolder(current) { return current ? current + '/Videos' : '/home/user/Videos' },
    async ChooseCookiesFile() { return '/home/user/cookies.txt' },
    async ChooseExecutable() { return '/usr/local/bin/yt-dlp' },
    async GetEngineStatus() { return engine },
    async InstallEngine() { return engine },
    async UpdateEngine() {
      for (const pct of [10, 40, 75, 100]) {
        await new Promise((r) => setTimeout(r, 300))
        emit('engine:progress', { component: 'yt-dlp', stage: pct === 100 ? 'done' : 'downloading', percent: pct, downloaded: pct * 300000, total: 30000000 })
      }
      return engine
    },
    async CheckEngineUpdates() { return engine },
    async GetUpdateStatus() { return { ...update } },
    async CheckForUpdates() { await new Promise((r) => setTimeout(r, 600)); return checkUpdate() },
    async DownloadUpdate() {
      update.busy = true
      for (const pct of [8, 30, 55, 80, 100]) {
        await new Promise((r) => setTimeout(r, 350))
        emit('update:progress', { stage: 'downloading', percent: pct, downloaded: pct * 210000, total: 21000000 })
      }
      emit('update:progress', { stage: 'verifying', percent: 100, downloaded: 0, total: 0 })
      await new Promise((r) => setTimeout(r, 400))
      Object.assign(update, { busy: false, ready: true })
      emit('update:progress', { stage: 'ready', percent: 100, downloaded: 0, total: 0 })
      return { ...update }
    },
    async RestartToUpdate() { location.reload() },
    async ListHistory(q) {
      let items = history.filter((h) => (!q.search || h.title.toLowerCase().includes(q.search.toLowerCase())) && (!q.platform || h.platform === q.platform) && (!q.status || h.status === q.status))
      const total = items.length
      items = items.slice(q.offset, q.offset + q.limit)
      return { items, total }
    },
    async HistoryPlatforms() { return [...new Set(history.map((h) => h.platform))] },
    async DeleteHistory(id) { const i = history.findIndex((h) => h.id === id); if (i >= 0) history.splice(i, 1) },
    async ClearHistory() { history.length = 0 },
    async GetDiagnostics() { return 'Dlyven 1.0.0-dev (browser mock)\nNo backend attached.' },
  }

  return {
    backend,
    on(name: string, cb: (...a: any[]) => void) {
      if (!listeners.has(name)) listeners.set(name, new Set())
      listeners.get(name)!.add(cb)
      return () => listeners.get(name)!.delete(cb)
    },
  }
}
