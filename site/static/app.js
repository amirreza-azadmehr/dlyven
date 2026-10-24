(() => {
  const root = document.documentElement
  const storedTheme = localStorage.getItem("dlyven-theme")
  const preferDark = window.matchMedia("(prefers-color-scheme: dark)").matches
  const theme = storedTheme || (preferDark ? "dark" : "light")
  root.setAttribute("data-theme", theme)

  const themeBtn = document.getElementById("theme-toggle")
  if (themeBtn) {
    themeBtn.addEventListener("click", () => {
      const next = root.getAttribute("data-theme") === "dark" ? "light" : "dark"
      root.setAttribute("data-theme", next)
      localStorage.setItem("dlyven-theme", next)
    })
  }

  const langSel = document.getElementById("lang-switch")
  if (langSel) {
    langSel.addEventListener("change", () => {
      localStorage.setItem("dlyven-lang", langSel.value)
      location.href = langSel.value
    })
  }

  const cfg = window.DLYVEN || {}
  if (cfg.page === "home" && cfg.lang === "en" && !sessionStorage.getItem("dlyven-lang-applied")) {
    sessionStorage.setItem("dlyven-lang-applied", "1")
    const saved = localStorage.getItem("dlyven-lang")
    const nav = (navigator.language || "en").toLowerCase()
    const map = [
      ["fa", "fa"], ["ar", "ar"], ["es", "es"], ["fr", "fr"], ["tr", "tr"],
      ["ru", "ru"], ["zh", "zh-CN"], ["de", "de"], ["pt", "pt-BR"],
    ]
    const hit = saved || (map.find(([p]) => nav.startsWith(p)) || [])[1]
    if (hit && hit !== "en") {
      const dest = document.querySelector(`#lang-switch option[data-code="${hit}"]`)
      if (dest) location.replace(dest.value)
    }
  }

  const owner = cfg.owner || "amirreza-azadmehr"
  const repo = cfg.repo || "dlyven"
  const notes = document.getElementById("notes")
  if (notes) {
    fetch(`https://api.github.com/repos/${owner}/${repo}/releases`)
      .then((r) => (r.ok ? r.json() : Promise.reject(r.status)))
      .then((list) => {
        notes.innerHTML = ""
        ;(list || []).forEach((rel) => {
          const art = document.createElement("article")
          const h = document.createElement("h2")
          const a = document.createElement("a")
          a.href = rel.html_url
          a.textContent = rel.name || rel.tag_name || "release"
          h.appendChild(a)
          art.appendChild(h)
          if (rel.published_at) {
            const d = document.createElement("p")
            d.className = "note"
            d.textContent = rel.published_at.slice(0, 10)
            art.appendChild(d)
          }
          const body = document.createElement("p")
          body.className = "note"
          body.textContent = (rel.body || "").replace(/\r\n/g, "\n").trim()
          art.appendChild(body)
          notes.appendChild(art)
        })
      })
      .catch(() => {
        notes.innerHTML = `<p><a href="https://github.com/${owner}/${repo}/releases">GitHub Releases</a></p>`
      })
  }

  const box = document.getElementById("releases")
  if (!box) return

  const os = /Win/i.test(navigator.userAgent) ? "windows"
    : /Mac/i.test(navigator.userAgent) ? "macos"
    : "linux"

  fetch(`https://api.github.com/repos/${owner}/${repo}/releases/latest`)
    .then((r) => (r.ok ? r.json() : Promise.reject(r.status)))
    .then((rel) => {
      const assets = rel.assets || []
      box.innerHTML = ""
      const h = document.createElement("p")
      h.textContent = rel.tag_name || rel.name || "latest"
      box.appendChild(h)
      if (!assets.length) {
        const p = document.createElement("p")
        p.textContent = cfg.noAssets || ""
        box.appendChild(p)
      }
      assets.forEach((a) => {
        const name = a.name.toLowerCase()
        // Per-file checksums and signatures; SHA256SUMS.txt is linked below the cards.
        if (/\.(sha256|asc)$/.test(name) || name === "sha256sums.txt") return
        const kind = name.includes("win") ? "windows"
          : name.includes("mac") || name.includes("darwin") ? "macos"
          : name.includes("linux") ? "linux"
          : "other"
        if (kind === "other") return
        const row = document.createElement("div")
        row.className = "card" + (kind === os ? " os" : "")
        const link = document.createElement("a")
        link.className = "btn btn-primary"
        link.href = a.browser_download_url
        link.textContent = a.name
        row.appendChild(link)
        const meta = document.createElement("p")
        meta.className = "note"
        meta.textContent = `${(a.size / 1048576).toFixed(1)} MB`
        row.appendChild(meta)
        box.appendChild(row)
      })
      const sums = assets.find((a) => a.name === "SHA256SUMS.txt")
      if (sums) {
        const s = document.createElement("p")
        s.innerHTML = `<a href="${sums.browser_download_url}">${sums.name}</a>`
        box.appendChild(s)
      }
    })
    .catch(() => {
      const a = document.createElement("a")
      a.className = "btn btn-primary"
      a.href = `https://github.com/${owner}/${repo}/releases/latest`
      a.textContent = cfg.releaseFallback || "GitHub Releases"
      box.appendChild(a)
    })
})()
