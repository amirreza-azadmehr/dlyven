// Turns the analysis result + the user's selection into backend DownloadRequests.
import type { DownloadRequest, Entry, Format, MediaInfo, Settings } from './types'
import type { Selection } from '@/store/analysis'

export function findFormat(info: MediaInfo, id: string | null): Format | undefined {
  if (!id) return undefined
  for (const g of info.videoGroups) {
    const f = g.formats.find((x) => x.id === id)
    if (f) return f
  }
  return info.audio.find((x) => x.id === id)
}

function baseRequest(info: MediaInfo, sel: Selection, s: Settings): DownloadRequest {
  return {
    url: info.url,
    title: info.title,
    uploader: info.uploader,
    thumbnail: info.thumbnail,
    platform: info.platform.id,
    duration: info.duration,
    uploadDate: info.uploadDate,
    mediaId: info.id,
    playlistTitle: '',
    playlistIndex: 0,
    playlistItem: 0,
    kind: sel.tab,
    formatId: '',
    hasAudio: false,
    maxHeight: 0,
    height: 0,
    fps: 0,
    container: sel.container,
    audioFormat: sel.audioFormat,
    audioBitrate: sel.audioBitrate,
    audioFormatId: '',
    subLangs: [],
    subMode: sel.subMode,
    subFormat: sel.subFormat,
    autoSubs: false,
    thumbnailUrl: '',
    thumbnailFormat: sel.thumbFormat,
    embedThumbnail: sel.embedThumbnail,
    embedMetadata: sel.embedMetadata,
    embedChapters: s.embedChapters,
    outputDir: sel.outputDir,
    filenameTemplate: s.filenameTemplate,
    qualityLabel: '',
    album: '',
  }
}

/** File-name-friendly quality label ("1080p60 HDR", "MP3 320k", …). Not localized on purpose. */
export function qualityLabel(info: MediaInfo, sel: Selection): string {
  switch (sel.tab) {
    case 'audio':
      return sel.audioFormat === 'flac' || sel.audioFormat === 'wav' || sel.audioBitrate === 'best'
        ? sel.audioFormat.toUpperCase()
        : `${sel.audioFormat.toUpperCase()} ${sel.audioBitrate}k`
    case 'subtitles':
      return 'subtitles'
    case 'thumbnail': {
      const t = info.thumbnails.find((x) => x.url === sel.thumbUrl)
      return t?.width ? `${t.width}x${t.height}` : 'thumbnail'
    }
    default: {
      const f = findFormat(info, sel.formatId)
      if (f) return f.label + (f.hdr ? ' HDR' : '')
      if (sel.maxHeight) return `${sel.maxHeight}p`
      return info.maxHeight ? `${info.maxHeight}p` : 'best'
    }
  }
}

function applyKind(req: DownloadRequest, info: MediaInfo, sel: Selection) {
  const subs = sel.subs.map((k) => k.split('|'))
  req.subLangs = [...new Set(subs.map(([lang]) => lang))]
  req.autoSubs = subs.some(([, a]) => a === 'a')
  switch (sel.tab) {
    case 'video': {
      const f = findFormat(info, sel.formatId)
      if (f && !info.isPlaylist) {
        req.formatId = f.id
        req.hasAudio = f.hasAudio
        req.height = Math.min(f.width || f.height, f.height) || f.height
        req.fps = f.fps
      } else {
        req.maxHeight = sel.maxHeight
        req.height = sel.maxHeight
      }
      break
    }
    case 'audio':
      req.audioFormatId = info.isPlaylist ? '' : sel.audioFormatId
      req.subLangs = []
      break
    case 'subtitles':
      req.subMode = 'file'
      break
    case 'thumbnail':
      req.thumbnailUrl = sel.thumbUrl || info.thumbnail
      req.subLangs = []
      break
  }
}

function entryRequest(base: DownloadRequest, info: MediaInfo, e: Entry): DownloadRequest {
  const r: DownloadRequest = {
    ...base,
    title: e.title,
    uploader: e.uploader || info.uploader,
    thumbnail: e.thumbnail || info.thumbnail,
    duration: e.duration,
    mediaId: e.id,
    playlistTitle: info.title,
    playlistIndex: e.index,
    subLangs: [],
  }
  if (e.url && e.url !== info.url) {
    r.url = e.url
  } else {
    r.url = info.url
    r.playlistItem = e.index
  }
  if (r.kind === 'thumbnail') r.thumbnailUrl = e.thumbnail
  return r
}

export function buildRequests(info: MediaInfo, sel: Selection, s: Settings): DownloadRequest[] {
  const base = baseRequest(info, sel, s)
  applyKind(base, info, sel)
  base.qualityLabel = qualityLabel(info, sel)

  // Cover art of a Spotify item/album is a single image, downloaded directly.
  if (info.matched && sel.tab === 'thumbnail') return [base]
  if (info.matched) {
    // Spotify: every track is matched to a public source and tagged with its metadata.
    const entries = info.entries.filter((e) => sel.entries.includes(e.index) || !info.isPlaylist)
    return entries
      .filter((e) => e.spotify)
      .map((e) => ({
        ...entryRequest({ ...base, kind: 'audio' }, info, e),
        url: '',
        playlistItem: 0,
        spotify: e.spotify,
        album: info.spotify?.kind === 'album' ? info.title : '',
        playlistTitle: info.isPlaylist ? info.title : '',
      }))
  }
  if (!info.isPlaylist) return [base]
  return info.entries.filter((e) => sel.entries.includes(e.index)).map((e) => entryRequest(base, info, e))
}

/** Values used for the live filename preview. */
export function previewVars(info: MediaInfo, sel: Selection): Record<string, string> {
  const first = info.isPlaylist ? info.entries.find((e) => sel.entries.includes(e.index)) : undefined
  const f = findFormat(info, sel.formatId)
  const track = first?.spotify ?? info.entries[0]?.spotify
  return {
    title: first?.title ?? info.title,
    uploader: first?.uploader || info.uploader,
    quality: qualityLabel(info, sel),
    resolution: f ? `${Math.min(f.width || f.height, f.height) || f.height}p` : sel.maxHeight ? `${sel.maxHeight}p` : '',
    fps: f?.fps ? `${Math.round(f.fps)}fps` : '',
    platform: info.platform.name,
    id: first?.id ?? info.id,
    date: info.uploadDate,
    playlist: info.isPlaylist ? info.title : '',
    index: first ? String(first.index).padStart(2, '0') : '',
    artist: track?.artists.join(', ') ?? '',
    album: track?.album || (info.spotify?.kind === 'album' ? info.title : ''),
    track: track?.trackNumber ? String(track.trackNumber).padStart(2, '0') : '',
  }
}

export function expectedExt(sel: Selection): string {
  switch (sel.tab) {
    case 'audio':
      return sel.audioFormat
    case 'subtitles':
      return sel.subFormat
    case 'thumbnail':
      return sel.thumbFormat === 'original' ? 'jpg' : sel.thumbFormat
    default:
      return sel.container
  }
}
