#!/usr/bin/env node
import { cpSync, mkdirSync, rmSync, writeFileSync } from "node:fs"
import { dirname, join } from "node:path"
import { fileURLToPath } from "node:url"
import { LANGS, STRINGS } from "./i18n.mjs"

const ROOT = dirname(fileURLToPath(import.meta.url))
const DIST = join(ROOT, "dist")
const BASE = (process.env.PAGES_BASE || "/dlyven/").replace(/\/?$/, "/")
const OWNER = process.env.PAGES_OWNER || "amirreza-azadmehr"
const REPO = process.env.PAGES_REPO || "dlyven"
const SITE = `https://${OWNER}.github.io${BASE}`
const GH = `https://github.com/${OWNER}/${REPO}`

const PAGES = [
  { id: "home", file: "index.html", path: "" },
  { id: "downloads", file: "downloads/index.html", path: "downloads/" },
  { id: "docs", file: "docs/index.html", path: "docs/" },
  { id: "changelog", file: "changelog/index.html", path: "changelog/" },
  { id: "contributing", file: "contributing/index.html", path: "contributing/" },
  { id: "license", file: "license/index.html", path: "license/" },
  { id: "youtube", file: "youtube-downloader/index.html", path: "youtube-downloader/" },
  { id: "instagram", file: "instagram-downloader/index.html", path: "instagram-downloader/" },
  { id: "tiktok", file: "tiktok-downloader/index.html", path: "tiktok-downloader/" },
  { id: "spotify", file: "spotify/index.html", path: "spotify/" },
]

const esc = (s) => String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]))

function prefix(code) {
  return code === "en" ? BASE : `${BASE}${code}/`
}
function abs(code, path) {
  return SITE.replace(/\/$/, "/") + (code === "en" ? path : `${code}/${path}`)
}

function hreflang(pagePath) {
  return LANGS.map((l) => `<link rel="alternate" hreflang="${l.code}" href="${abs(l.code, pagePath)}">`).join("\n    ")
    + `\n    <link rel="alternate" hreflang="x-default" href="${abs("en", pagePath)}">`
}

function layout({ lang, page, title, description, canonical, body, jsonLd }) {
  const t = STRINGS[lang.code]
  const dir = lang.dir
  const options = LANGS.map((l) => {
    const href = `${prefix(l.code)}${page.path}`
    return `<option value="${href}" data-code="${l.code}"${l.code === lang.code ? " selected" : ""}>${esc(l.name)}</option>`
  }).join("")
  const nav = [
    ["home", t.nav.home, ""],
    ["downloads", t.nav.downloads, "downloads/"],
    ["docs", t.nav.docs, "docs/"],
    ["changelog", t.nav.changelog, "changelog/"],
    ["contributing", t.nav.contributing, "contributing/"],
    ["license", t.nav.license, "license/"],
  ].map(([id, label, p]) => {
    const href = `${prefix(lang.code)}${p}`
    const cur = page.id === id ? ' aria-current="page"' : ""
    return `<a href="${href}"${cur}>${esc(label)}</a>`
  }).join("")

  return `<!doctype html>
<html lang="${lang.code}" dir="${dir}">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>${esc(title)}</title>
    <meta name="description" content="${esc(description)}">
    <link rel="canonical" href="${canonical}">
    ${hreflang(page.path)}
    <meta property="og:type" content="website">
    <meta property="og:title" content="${esc(title)}">
    <meta property="og:description" content="${esc(description)}">
    <meta property="og:url" content="${canonical}">
    <meta property="og:image" content="${SITE}icon.png">
    <meta name="twitter:card" content="summary">
    <meta name="twitter:title" content="${esc(title)}">
    <meta name="twitter:description" content="${esc(description)}">
    <meta name="twitter:image" content="${SITE}icon.png">
    <link rel="icon" href="${BASE}icon.png">
    <link rel="apple-touch-icon" href="${BASE}icon.png">
    <link rel="stylesheet" href="${BASE}styles.css">
    ${jsonLd ? `<script type="application/ld+json">${jsonLd}</script>` : ""}
</head>
<body>
    <a class="skip" href="#main">Skip to content</a>
    <header>
      <div class="wrap bar">
        <a class="brand" href="${prefix(lang.code)}"><img src="${BASE}icon.png" width="32" height="32" alt=""> Dlyven</a>
        <nav>${nav}</nav>
        <div class="tools">
          <label class="lang-switch">
            <span class="lang-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="9"/>
                <path d="M3 12h18"/>
                <path d="M12 3a14 14 0 0 1 0 18"/>
                <path d="M12 3a14 14 0 0 0 0 18"/>
              </svg>
            </span>
            <select id="lang-switch" aria-label="${esc(t.nav.lang)}">${options}</select>
          </label>
          <button type="button" class="btn btn-ghost" id="theme-toggle">${esc(t.nav.theme)}</button>
        </div>
      </div>
    </header>
    <main id="main" class="wrap">${body}</main>
    <footer>
      <div class="wrap">
        <p>${esc(t.footer.credit)} · <a href="${GH}">GitHub</a> · <a href="https://tr.linkedin.com/in/amirreza-azadmehr">LinkedIn</a> · <a href="https://t.me/amir_reza_azadmehr">Telegram</a></p>
        <p class="legal">${esc(t.footer.notice)}</p>
      </div>
    </footer>
    <script>window.DLYVEN=${JSON.stringify({
      owner: OWNER, repo: REPO, lang: lang.code, page: page.id,
      releaseFallback: t.downloads.fallback, noAssets: t.downloads.noAssets,
    })}</script>
    <script src="${BASE}app.js" defer></script>
</body>
</html>
`
}

function brandRow() {
  return [
    ["youtube", "YouTube"],
    ["instagram", "Instagram"],
    ["tiktok", "TikTok"],
    ["spotify", "Spotify"],
    ["x", "X"],
    ["facebook", "Facebook"],
    ["soundcloud", "SoundCloud"],
    ["twitch", "Twitch"],
  ].map(([file, alt]) => `<img src="${BASE}images/brands/${file}.svg" alt="${alt}" width="28" height="28">`).join("")
}

function home(lang, t) {
  const feats = t.features.map((f) => `<article class="card"><h3>${esc(f.title)}</h3><p>${esc(f.body)}</p></article>`).join("")
  return `
<section class="hero">
  <h1>${esc(t.hero.title)}</h1>
  <p class="lead">${esc(t.hero.tagline)}</p>
  <div class="cta">
    <a class="btn btn-primary" href="${prefix(lang.code)}downloads/">${esc(t.hero.download)}</a>
    <a class="btn" href="${GH}">${esc(t.hero.github)}</a>
  </div>
  <div class="shots">
    <a href="${BASE}images/home.png"><img src="${BASE}images/home.png" alt="Home" width="1180" height="780"></a>
    <a href="${BASE}images/formats.png"><img src="${BASE}images/formats.png" alt="Formats" width="1180" height="780"></a>
    <a href="${BASE}images/queue.png"><img src="${BASE}images/queue.png" alt="Queue" width="1180" height="780"></a>
  </div>
  <div class="brands">${brandRow()}</div>
  <p class="note">${esc(t.footer.notice)}</p>
  <p class="note">${esc(t.hero.disclaimer)}</p>
</section>
<section class="grid grid-3">${feats}</section>
<section>
  <h2>${esc(t.sites.title)}</h2>
  <p>${esc(t.sites.body)}</p>
  <p>
    <a href="${prefix(lang.code)}youtube-downloader/">YouTube</a> ·
    <a href="${prefix(lang.code)}instagram-downloader/">Instagram</a> ·
    <a href="${prefix(lang.code)}tiktok-downloader/">TikTok</a> ·
    <a href="${prefix(lang.code)}spotify/">Spotify</a>
  </p>
</section>`
}

function downloads(t) {
  return `<article>
  <h1>${esc(t.downloads.title)}</h1>
  <p>${esc(t.downloads.body)}</p>
  <p class="note">${esc(t.downloads.checksums)}</p>
  <div id="releases" class="grid grid-3"></div>
  <p><a class="btn" href="${GH}/releases/latest">${esc(t.downloads.fallback)}</a></p>
</article>`
}

function docs(t) {
  const faq = t.docs.faq.map((x) => `<h3>${esc(x.q)}</h3><p>${esc(x.a)}</p>`).join("")
  return `<article>
  <h1>${esc(t.docs.title)}</h1>
  <p>${esc(t.docs.install)}</p>
  <p>${esc(t.docs.usage)}</p>
  <p>${esc(t.docs.config)}</p>
  <h2>${esc(t.docs.faqTitle)}</h2>${faq}
</article>`
}

function simple(title, body) {
  return `<article><h1>${esc(title)}</h1><p>${esc(body)}</p></article>`
}

function changelog(t) {
  return `<article>
  <h1>${esc(t.changelog.title)}</h1>
  <p>${esc(t.changelog.body)}</p>
  <p><a href="${GH}/releases">${esc(t.changelog.link)}</a></p>
  <div id="notes" class="notes"></div>
</article>`
}

function platform(block) {
  const steps = `<ol>${block.steps.map((s) => `<li>${esc(s)}</li>`).join("")}</ol>`
  return `<article>
  <h1>${esc(block.h1)}</h1>
  <p>${esc(block.intro)}</p>
  ${steps}
  <p class="note">${esc(block.limits)}</p>
</article>`
}

function jsonLdHome() {
  return JSON.stringify({
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "Dlyven",
    applicationCategory: "MultimediaApplication",
    operatingSystem: "Windows, macOS, Linux",
    offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
    downloadUrl: `${GH}/releases/latest`,
    softwareVersion: "1.0.0",
    license: "https://www.gnu.org/licenses/gpl-3.0.html",
    url: SITE,
  })
}

function writePage(lang, page) {
  const t = STRINGS[lang.code]
  const canonical = abs(lang.code, page.path)
  let title = t.meta.homeTitle
  let description = t.meta.homeDesc
  let body = ""
  let jsonLd = ""
  switch (page.id) {
    case "home":
      body = home(lang, t); jsonLd = jsonLdHome(); break
    case "downloads":
      title = t.meta.downloadsTitle; description = t.meta.downloadsDesc; body = downloads(t); break
    case "docs":
      title = t.meta.docsTitle; description = t.meta.docsDesc; body = docs(t); break
    case "changelog":
      title = t.meta.changelogTitle; description = t.changelog.body; body = changelog(t); break
    case "contributing":
      title = t.meta.contributingTitle; description = t.contributing.body; body = simple(t.contributing.title, t.contributing.body); break
    case "license":
      title = t.meta.licenseTitle; description = t.license.body; body = simple(t.license.title, t.license.body); break
    case "youtube":
      title = t.meta.youtubeTitle; description = t.meta.youtubeDesc; body = platform(t.youtube); break
    case "instagram":
      title = t.meta.instagramTitle; description = t.meta.instagramDesc; body = platform(t.instagram); break
    case "tiktok":
      title = t.meta.tiktokTitle; description = t.meta.tiktokDesc; body = platform(t.tiktok); break
    case "spotify":
      title = t.meta.spotifyTitle; description = t.meta.spotifyDesc; body = platform(t.spotify); break
  }
  const html = layout({ lang, page, title, description, canonical, body, jsonLd })
  const out = join(DIST, lang.code === "en" ? page.file : join(lang.code, page.file))
  mkdirSync(dirname(out), { recursive: true })
  writeFileSync(out, html)
}

rmSync(DIST, { recursive: true, force: true })
mkdirSync(DIST, { recursive: true })
cpSync(join(ROOT, "static"), DIST, { recursive: true })
const imgDir = join(DIST, "images")
mkdirSync(imgDir, { recursive: true })
for (const name of ["home.png", "formats.png", "queue.png"]) {
  cpSync(join(ROOT, "..", "docs", "images", name), join(imgDir, name))
}
cpSync(join(ROOT, "..", "docs", "images", "brands"), join(imgDir, "brands"), { recursive: true })

for (const lang of LANGS) {
  for (const page of PAGES) writePage(lang, page)
}

const urls = []
for (const lang of LANGS) {
  for (const page of PAGES) urls.push(`  <url><loc>${abs(lang.code, page.path)}</loc></url>`)
}
writeFileSync(join(DIST, "sitemap.xml"), `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls.join("\n")}
</urlset>
`)
writeFileSync(join(DIST, "robots.txt"), `User-agent: *
Allow: /
Sitemap: ${SITE}sitemap.xml
`)
writeFileSync(join(DIST, "404.html"), layout({
  lang: LANGS[0],
  page: PAGES[0],
  title: "Not found — Dlyven",
  description: "Page not found",
  canonical: `${SITE}404.html`,
  body: `<article><h1>Not found</h1><p><a href="${BASE}">Home</a></p></article>`,
  jsonLd: "",
}))

console.log(`Built ${LANGS.length} languages × ${PAGES.length} pages → site/dist (base ${BASE})`)
