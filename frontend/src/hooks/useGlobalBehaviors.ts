import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { rt } from '@/lib/api'
import { detectPlatform, findUrls } from '@/lib/platforms'
import { playChime } from '@/lib/sound'
import type { Job } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { useUpdate } from '@/store/update'
import { DOWNLOAD_EVENT } from '@/pages/home/DownloadBar'
import { errorHint, errorTitle } from '@/components/ErrorPanel'

const isTyping = (el: EventTarget | null) =>
  el instanceof HTMLElement && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)

/**
 * Keyboard shortcuts:
 *   Ctrl/Cmd+V     paste & analyze (when not typing in a field)
 *   Ctrl/Cmd+Enter start the download on the result screen
 *   Ctrl/Cmd+,     open settings
 *   Esc            cancel analysis / go back from the result view
 */
export function useShortcuts() {
  useEffect(() => {
    const onKey = async (e: KeyboardEvent) => {
      const mod = e.ctrlKey || e.metaKey
      const ui = useUI.getState()
      const an = useAnalysis.getState()
      if (mod && e.key === ',') {
        e.preventDefault()
        ui.navigate('settings')
        return
      }
      if (mod && e.key === 'Enter') {
        if (ui.route === 'home' && an.status === 'ready') {
          e.preventDefault()
          window.dispatchEvent(new Event(DOWNLOAD_EVENT))
        }
        return
      }
      if (mod && e.key.toLowerCase() === 'v' && !isTyping(e.target)) {
        e.preventDefault()
        const text = (await rt.ClipboardGetText().catch(() => '')) ?? ''
        const url = findUrls(text)[0]
        if (!url) return
        ui.navigate('home')
        an.setUrl(url)
        void an.analyze(url)
        return
      }
      if (e.key === 'Escape' && !document.querySelector('[role="dialog"]')) {
        if (an.status === 'analyzing') an.cancel()
        else if (ui.route === 'home' && (an.status === 'ready' || an.status === 'error')) an.reset()
        else if (ui.route !== 'home') ui.navigate('home')
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])
}

/** When the window regains focus, offer to analyze a newly copied supported link. */
export function useClipboardWatch() {
  const { t } = useTranslation()
  const last = useRef('')
  const enabled = useSettings((s) => s.settings?.clipboardWatch ?? false)

  useEffect(() => {
    if (!enabled) return
    const check = async () => {
      const text = (await rt.ClipboardGetText().catch(() => '')) ?? ''
      const url = findUrls(text)[0]
      if (!url || url === last.current) return
      last.current = url
      const registry = useSettings.getState().app?.platforms ?? []
      const p = detectPlatform(url, registry)
      const an = useAnalysis.getState()
      if (!p.valid || an.url.trim() === url || an.status === 'analyzing') return
      useUI.getState().toast({
        kind: 'info',
        title: t('home.clipboardFound'),
        body: url,
        duration: 8000,
        action: {
          label: t('home.clipboardAnalyze'),
          run: () => {
            useUI.getState().navigate('home')
            an.setUrl(url)
            void an.analyze(url)
          },
        },
      })
    }
    // Ignore whatever is in the clipboard when the feature is switched on.
    void rt.ClipboardGetText().then((x) => (last.current = findUrls(x ?? '')[0] ?? '')).catch(() => {})
    window.addEventListener('focus', check)
    return () => window.removeEventListener('focus', check)
  }, [enabled, t])
}

/** Toast + sound when a download finishes (the backend sends the native notification). */
export function useFinishFeedback() {
  const { t } = useTranslation()
  useEffect(() => {
    return rt.EventsOn('download:finished', (j: Job) => {
      const s = useSettings.getState().settings
      const ui = useUI.getState()
      if (j.status === 'done') {
        if (s?.playSound) playChime()
        ui.toast({ kind: 'success', title: t('downloads.doneToast'), body: j.request.title })
      } else if (j.status === 'failed') {
        ui.toast({
          kind: 'error',
          title: `${t('downloads.failedToast')}: ${errorTitle(t, j.errorCode)}`,
          body: errorHint(t, j.errorCode),
          action: { label: t('result.viewDownloads'), run: () => ui.navigate('downloads') },
        })
      }
    })
  }, [t])
}

let updateNoticeShown = false

/** One toast per session when a new Dlyven is out. Settings → Updates does the rest. */
export function useUpdateNotice() {
  const { t } = useTranslation()
  useEffect(() => {
    return useUpdate.subscribe(({ status }) => {
      if (updateNoticeShown || !status?.available) return
      updateNoticeShown = true
      const ui = useUI.getState()
      if (ui.route === 'settings') return
      ui.toast({
        kind: 'info',
        title: t('updates.toastTitle', { version: status.latest }),
        body: t('updates.toastBody'),
        action: { label: t('updates.toastAction'), run: () => ui.navigate('settings', 'updates') },
        duration: 12000,
      })
    })
  }, [t])
}
