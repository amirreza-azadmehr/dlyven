import { AnimatePresence, MotionConfig, motion } from 'framer-motion'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { applyDevPreview } from '@/lib/devPreview'
import { useDownloads } from '@/store/downloads'
import { useEngine } from '@/store/engine'
import { useSettings } from '@/store/settings'
import { useUI } from '@/store/ui'
import { useClipboardWatch, useFinishFeedback, useShortcuts } from '@/hooks/useGlobalBehaviors'
import { Sidebar } from '@/components/layout/Sidebar'
import { TitleBar } from '@/components/layout/TitleBar'
import { Toaster } from '@/components/ui/Feedback'
import { TooltipProvider } from '@/components/ui/Tooltip'
import { HomePage } from '@/pages/home/HomePage'
import { DownloadsPage } from '@/pages/downloads/DownloadsPage'
import { HistoryPage } from '@/pages/history/HistoryPage'
import { SettingsPage } from '@/pages/settings/SettingsPage'

export default function App() {
  const { i18n } = useTranslation()
  const [ready, setReady] = useState(false)
  const route = useUI((s) => s.route)
  const loadSettings = useSettings((s) => s.load)

  useEffect(() => {
    let offs: Array<() => void> = []
    let disposed = false
    ;(async () => {
      try {
        await loadSettings()
      } catch (e) {
        console.error('init failed', e)
      }
      if (!disposed) {
        setReady(true)
        applyDevPreview()
      }
      // Engine status used to exec yt-dlp --version (45s hang on macOS) before Home
      // painted. Subscribe in the background so search is usable immediately.
      const subs = await Promise.all([useDownloads.getState().init(), useEngine.getState().init()]).catch((e) => {
        console.error('init failed', e)
        return [] as Array<() => void>
      })
      if (disposed) return subs.forEach((off) => off())
      offs = subs
    })()
    return () => {
      disposed = true
      offs.forEach((off) => off())
    }
  }, [loadSettings])

  // Never let a drop outside the Home drop zone navigate the webview to the dropped file.
  useEffect(() => {
    const block = (e: DragEvent) => e.preventDefault()
    window.addEventListener('dragover', block)
    window.addEventListener('drop', block)
    return () => {
      window.removeEventListener('dragover', block)
      window.removeEventListener('drop', block)
    }
  }, [])

  useShortcuts()
  useClipboardWatch()
  useFinishFeedback()

  const dir = i18n.dir() === 'rtl' ? -1 : 1

  return (
    <MotionConfig reducedMotion="user">
      <TooltipProvider>
        <div className="ambient flex h-full" style={{ backgroundColor: 'rgb(var(--bg) / var(--shell-alpha))' }}>
          <Sidebar />
          <div className="flex min-w-0 flex-1 flex-col">
            <TitleBar />
            <main className="relative min-h-0 flex-1 overflow-y-auto overflow-x-hidden" id="main" tabIndex={-1}>
              {ready && (
                <AnimatePresence mode="wait" initial={false}>
                  <motion.div
                    key={route}
                    className="h-full"
                    initial={{ opacity: 0, x: 10 * dir }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: -10 * dir }}
                    transition={{ duration: 0.16, ease: 'easeOut' }}
                  >
                    {route === 'home' && <HomePage />}
                    {route === 'downloads' && <DownloadsPage />}
                    {route === 'history' && <HistoryPage />}
                    {route === 'settings' && <SettingsPage />}
                  </motion.div>
                </AnimatePresence>
              )}
            </main>
          </div>
        </div>
        <Toaster />
      </TooltipProvider>
    </MotionConfig>
  )
}
