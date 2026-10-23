// Browser-mock only: open a specific screen via URL params, e.g.
//   http://localhost:5173/?lang=ckb&route=home&url=https://youtu.be/demo
//   ?route=downloads&seed=1      ?route=settings&theme=light
import { isWails } from './api'
import { useAnalysis } from '@/store/analysis'
import { useUI, type Route } from '@/store/ui'

export function applyDevPreview() {
  if (isWails) return
  const q = new URLSearchParams(location.search)
  const route = q.get('route') as Route | null
  if (route) useUI.getState().navigate(route)
  const url = q.get('url')
  if (url) {
    useAnalysis.getState().setUrl(url)
    void useAnalysis.getState().analyze(url)
  }
}
