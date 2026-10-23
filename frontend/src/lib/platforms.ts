// Instant client-side platform detection for the live badge while typing. It uses the
// registry served by the backend (App.GetAppInfo().platforms), so platforms are defined once in Go.
import type { PlatformDef, PlatformInfo } from './types'

const EMPTY: PlatformInfo = { id: '', name: '', color: '', icon: '', valid: false, known: false, audioOnly: false, matched: false, url: '' }

export function normalizeUrl(raw: string): string | null {
  let s = raw.trim().replace(/^[<"'`(\[‎‏]+|[>"'`)\]‎‏]+$/g, '')
  if (!s) return null
  if (/^spotify:/i.test(s)) return s
  if (!/^[a-z][a-z0-9+.-]*:\/\//i.test(s)) {
    if (!/^([a-z0-9-]+\.)+[a-z]{2,}(\/|$)/i.test(s)) return null
    s = 'https://' + s
  }
  try {
    const u = new URL(s)
    if (!/^https?:$/.test(u.protocol) || !u.hostname.includes('.')) return null
    return u.toString()
  } catch {
    return null
  }
}

export function detectPlatform(raw: string, registry: PlatformDef[]): PlatformInfo {
  const url = normalizeUrl(raw)
  if (!url) return EMPTY
  const spotify = registry.find((p) => p.id === 'spotify')
  if (/^spotify:/i.test(url) && spotify) return { ...spotify, valid: true, known: true, url }
  const host = new URL(url).hostname.toLowerCase().replace(/^www\./, '')
  const def = registry.find((p) => p.hosts.some((h) => host === h || host.endsWith('.' + h)))
  if (def) return { ...def, valid: true, known: true, url }
  return { id: 'generic', name: host, color: '#8B93A7', icon: 'globe', valid: true, known: false, audioOnly: false, matched: false, url }
}

/** Finds all links in pasted/dropped text (client-side quick check; backend re-validates). */
export function findUrls(text: string): string[] {
  const out: string[] = []
  for (const m of text.match(/https?:\/\/[^\s<>"'`]+|spotify:[a-z]+:[A-Za-z0-9]+/gi) ?? []) {
    const u = normalizeUrl(m.replace(/[.,;:!?)\]}»،]+$/, ''))
    if (u && !out.includes(u)) out.push(u)
  }
  return out
}
