# Dlyven

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![CI](https://github.com/amirreza-azadmehr/dlyven/actions/workflows/ci.yml/badge.svg)](https://github.com/amirreza-azadmehr/dlyven/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/amirreza-azadmehr/dlyven)](https://github.com/amirreza-azadmehr/dlyven/releases)
[![Downloads](https://img.shields.io/github/downloads/amirreza-azadmehr/dlyven/total)](https://github.com/amirreza-azadmehr/dlyven/releases)

<p align="center">
  <img src="docs/images/home.png" alt="Home" width="880">
</p>
<p align="center">
  <img src="docs/images/formats.png" alt="Formats" width="880">
</p>
<p align="center">
  <img src="docs/images/queue.png" alt="Queue" width="880">
</p>

Paste a link, pick a quality, hit download. That’s the whole app.

Desktop, open source, Windows / macOS / Linux. Under the hood it’s [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — Dlyven is just a nicer window around them.

**Site:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/)

**Languages:** [English](README.md) · [فارسی](docs/i18n/README.fa.md) · [Español](docs/i18n/README.es.md) · [Français](docs/i18n/README.fr.md) · [Türkçe](docs/i18n/README.tr.md) · [العربية](docs/i18n/README.ar.md) · [Русский](docs/i18n/README.ru.md) · [简体中文](docs/i18n/README.zh-CN.md) · [Deutsch](docs/i18n/README.de.md) · [Português (Brasil)](docs/i18n/README.pt-BR.md) · [کوردی](docs/i18n/README.ckb.md)

<p align="center">
  <img src="docs/images/brands/youtube.svg" alt="YouTube" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/instagram.svg" alt="Instagram" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/tiktok.svg" alt="TikTok" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/spotify.svg" alt="Spotify" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/x.svg" alt="X" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/facebook.svg" alt="Facebook" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/soundcloud.svg" alt="SoundCloud" height="28">
  &nbsp;&nbsp;
  <img src="docs/images/brands/twitch.svg" alt="Twitch" height="28">
</p>

<p align="center">Those logos aren’t ours. YouTube, Instagram, TikTok, Spotify and the rest belong to them — we’re just a downloader.</p>

> Only grab stuff you own or have permission to download.

UI talks **English, فارسی, Türkçe, کوردی (سۆرانی)** and flips properly for RTL. Stack: Go + Wails, React, TypeScript, Tailwind.

---

## Install

1. Open the [latest release](https://github.com/amirreza-azadmehr/dlyven/releases/latest).
2. Download **one** file:

| Your computer | File |
| --- | --- |
| Windows | `Dlyven-windows-amd64.exe` |
| Mac (Intel or Apple silicon) | `Dlyven-macos-universal.zip` |
| Linux x64 | `Dlyven-linux-amd64` |
| Linux ARM | `Dlyven-linux-arm64` |

3. Open it. Windows: double-click the `.exe`. Linux: `chmod +x Dlyven-linux-*` then run it. Mac: unzip and open `Dlyven.app`.
4. First launch needs internet — it grabs yt-dlp and ffmpeg by itself. After that, paste a link and download.

That’s it. No installer, no extra programs.

Official GitHub builds are meant to be **signed** (Apple notarized on Mac, Authenticode on Windows) so Gatekeeper / SmartScreen stay quiet. That only happens after the signing certs are in the repo secrets — until then macOS may still say it “could not verify” the app.

## What it does

- Paste a link (or hit <kbd>Ctrl/⌘ V</kbd>, or drop a URL / a `.txt` of links).
- It figures out the site, shows the title and thumbnail, then every quality it can find — video, audio, subtitles, thumbnail.
- Presets if you don’t want to think: Best, 4K, 1080p, MP3, that kind of thing.
- Queue with pause / resume / cancel / retry. Close the app, come back, unfinished stuff is still there.
- History, search, pick a folder, name files however you like.
- Spotify: we **don’t** pull audio from Spotify. We read the public title/artist, find a public match, and tag the file. Podcasts aren’t a thing here.
- Private posts: Settings → Login & cookies, if you’re signed in in the browser.
- First launch grabs yt-dlp, ffmpeg and Deno on its own. After that you’re good.

Shortcuts: <kbd>Ctrl/⌘ V</kbd> paste · <kbd>Ctrl/⌘ Enter</kbd> download · <kbd>Ctrl/⌘ ,</kbd> settings · <kbd>Esc</kbd> back.

## Sites

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… and [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Unknown links still work; they just show up as “Other site”.

**YouTube** — paste the link, pick video or MP3, download. Playlists get a checklist.

**Instagram** — reel or post URL. Private ones need you to be logged in (cookies).

**TikTok** — paste it, video or audio. Short links are fine.

**Spotify** — metadata only. The actual audio is a public match, not Spotify’s file.

## FAQ

**Is this an official YouTube / Instagram / TikTok app?**  
Nope. Just a desktop UI on top of yt-dlp.

**Can I rip songs straight from Spotify?**  
No. We only match a public recording and tag it.

**Does it phone home?**  
No.

**Where’s my stuff?**  
Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Roadmap

- Apple notarization + Windows Authenticode once cert secrets are set
- More UI languages
- Optional tray / start-on-login
- Better playlist UX for very large lists

Ideas: [GitHub Issues](https://github.com/amirreza-azadmehr/dlyven/issues).

## Requirements

| Tool | Version |
| --- | --- |
| Go | 1.25+ |
| Node.js | 20+ (npm 10+) |
| Wails CLI | v2.10.2: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2` |

Platform prerequisites (run `wails doctor` to verify):

- **Windows 10/11**: nothing else. WebView2 ships with Windows 11, and builds embed the bootstrapper for Windows 10.
- **macOS 11+**: Xcode Command Line Tools (`xcode-select --install`).
- **Linux**: `build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev` (Ubuntu 24.04+/Debian 13+) or `libwebkit2gtk-4.0-dev` on older distros.

You don’t need yt-dlp or ffmpeg installed first. Dlyven grabs them on first launch, or uses the ones already on your machine.

## Development

Push real app changes to `main` and GitHub tags the next patch, builds Windows/macOS/Linux, and writes the release notes from the commits. Docs-only pushes skip that. Put `[skip release]` on a commit if you don’t want a build.

```bash
cd frontend && npm install && cd ..
wails dev
```

`wails dev` runs the Go backend with hot-reloading Vite. Logs go to the data folder (below); set `DLYVEN_DEBUG=1` to mirror them to the terminal.

**UI-only mode.** `cd frontend && npm run dev` opens the UI in a normal browser with a built-in **mock backend** (fake analysis, a simulated queue with progress, history). This is the fastest way to iterate on design. URL parameters open any state directly:

```
http://localhost:5173/?lang=fa&theme=dark                     # Persian, dark
http://localhost:5173/?lang=ckb&url=https://youtu.be/demo     # Kurdish, result view
http://localhost:5173/?url=https://open.spotify.com/album/x   # Spotify (matched) flow
http://localhost:5173/?url=https://instagram.com/p/private-error  # error state
http://localhost:5173/?route=downloads&seed=1                 # queue with sample jobs
http://localhost:5173/?route=settings&theme=light&accent=rose&digits=1
```

Data folder (settings, history, logs, managed binaries):

| OS | Path |
| --- | --- |
| Windows | `%AppData%\Dlyven` |
| macOS | `~/Library/Application Support/Dlyven` |
| Linux | `~/.config/Dlyven` |

## Building

Outputs go to `build/bin/`. Stamp a version with `VERSION=1.2.3` (bash) or `-Version 1.2.3` (PowerShell).

### Windows (portable `.exe`)

On Windows (PowerShell):

```powershell
powershell -ExecutionPolicy Bypass -File scripts\build.ps1
powershell -ExecutionPolicy Bypass -File scripts\build.ps1 -Installer   # optional NSIS setup (needs makensis)
```

From **Linux**, cross-compile a 64-bit Windows GUI exe (needs MinGW or llvm-mingw on `PATH`, e.g. `~/.local/opt/llvm-mingw`):

```bash
./scripts/build.sh windows
# → build/bin/Dlyven.exe
```

WebView2 is embedded (`-webview2 embed`), so the exe also runs on machines that do not already have it.

### Linux

```bash
./scripts/build.sh linux
# → build/bin/Dlyven
```

The script detects WebKitGTK 4.1 vs 4.0 and adds `-tags webkit2_41` on Ubuntu 24.04+ / Debian 13+.

### macOS (Intel + Apple Silicon)

Must run **on macOS**. Wails cannot cross-compile to Darwin from Linux or Windows. One universal `.app` covers Intel and Apple silicon:

```bash
./scripts/build.sh darwin
# → build/bin/Dlyven.app
```

### Wails directly

```bash
wails build -clean -trimpath
# Linux (newer distros):  wails build -clean -trimpath -tags webkit2_41
# Windows:                wails build -trimpath -platform windows/amd64 -webview2 embed
```

Release builds sign when GitHub secrets are present (`MACOS_CERTIFICATE_P12`, `APPLE_API_KEY` / `APPLE_API_KEY_ID` / `APPLE_API_ISSUER`, `WIN_CERTIFICATE_PFX`). Without those, Apple Gatekeeper and Windows SmartScreen will warn.

## Testing

```bash
go test ./internal/...                      # unit tests (URL detection, format parsing/grouping,
                                            # filename sanitizing, progress parsing, checksums,
                                            # error classification, settings, history)
go test -tags smoke -v ./internal/smoke     # opt-in end-to-end test against the real internet:
                                            # installs the engine into a temp dir, analyzes and
                                            # downloads a public video, and runs a Spotify match
cd frontend
npm run typecheck
npm run i18n:check                          # every locale has every key + same {{placeholders}}
npm run fonts:check                         # bundled fonts contain every Sorani/Persian glyph
```

`npm run build` runs the locale check and typecheck automatically.

## Project structure

```
.
├── main.go                     # Wails window options (frameless, Mica/vibrancy, single instance)
├── app.go                      # Methods bound to the frontend (window.go.main.App.*)
├── wails.json
├── build/appicon.png           # App icon (Wails generates .ico/.icns from it)
├── scripts/build.sh            # macOS / Linux builds
├── scripts/build.ps1           # Windows builds (+ optional NSIS installer)
├── internal/
│   ├── appinfo/                # Name, ID, version, data dir: rename the app here
│   ├── errs/                   # Stable error codes + yt-dlp output classifier
│   ├── platform/               # URL → platform registry; Spotify metadata + matching
│   ├── settings/               # Settings model, validation, atomic JSON persistence
│   ├── store/                  # Download history (SQLite via modernc.org/sqlite, no CGO)
│   ├── engine/                 # Install/update/locate yt-dlp, ffmpeg/ffprobe, Deno (+ SHA-256)
│   ├── extractor/              # Analyze(url) → MediaInfo (formats grouped, highlights, subs…)
│   ├── downloader/             # Queue, workers, args, progress parsing, pause/resume/cancel/retry
│   ├── media/                  # ffmpeg tagging (Spotify metadata + cover) and image conversion
│   ├── filename/               # Template rendering, Unicode-safe sanitizing, unique names
│   ├── logx/                   # JSON logging to a size-rotated file
│   ├── sysutil/                # Hidden subprocesses, process-tree kill, reveal in folder, notify
│   └── smoke/                  # Opt-in end-to-end tests (build tag: smoke)
└── frontend/
    ├── index.html, vite.config.ts, tailwind.config.js, postcss.config.js, tsconfig.json
    ├── scripts/check-locales.mjs, scripts/check-fonts.mjs
    └── src/
        ├── main.tsx, App.tsx, index.css          # Shell, design tokens, themes, accents
        ├── config/app.ts                         # Frontend app name
        ├── i18n/index.ts                         # Language registry (the one place to add a language)
        ├── locales/{en,fa,tr,ckb}.json
        ├── lib/        api.ts (typed bridge + mock fallback), mock.ts, types.ts, format.ts (Intl),
        │               requests.ts (selection → DownloadRequest), platforms.ts, sound.ts, …
        ├── store/      settings, analysis, downloads, engine, ui (Zustand)
        ├── hooks/      shortcuts, clipboard watch, finish feedback
        ├── components/ ui/ (Button, Select, Switch, Segmented, Dialog, Toaster, Progress…),
        │               layout/ (TitleBar, Sidebar), Platform, ErrorPanel, EngineBanner, Illustrations
        └── pages/      home/ (hero input, result view, tabs, batch import), downloads/, history/, settings/
```

## How it works

```
Paste URL ─▶ App.Analyze ─▶ extractor ─▶ yt-dlp --dump-single-json --flat-playlist
                                │              (or Spotify embed metadata)
                                ▼
                     MediaInfo (formats grouped, highlights, subtitles, entries)
                                │
User picks format/folder ─▶ App.Enqueue ─▶ downloader queue (N workers)
                                                │  yt-dlp --progress-template … --newline
                                                ▼
                            "download:update" events (≈10 Hz per job) ─▶ UI
                                                │
                     merge/convert (ffmpeg) ─▶ done ─▶ history (SQLite) + notification
```

- **Never blocks the UI.** Every bound method runs on its own goroutine; downloads report through Wails events (`download:update`, `download:removed`, `download:finished`, `engine:progress`, `engine:status`). Progress events are throttled to about 10 Hz per job; status changes are sent immediately.
- **Progress parsing.** yt-dlp prints machine-readable lines via `--progress-template` (`DLYVEN|status|downloaded|total|estimate|speed|eta|format_id|second_format_id` and `DLYVENPP|status|postprocessor`). Merged video+audio downloads report as one combined 0–100%.
- **Pause / resume / cancel** are context-based. The whole process tree (yt-dlp + ffmpeg) is killed: a process group on Unix, `taskkill /T` on Windows. Partial files live in `<destination>/.dlyven-partial/<job>/` (same disk, so the final move is instant). Pause keeps them and yt-dlp resumes with `--continue`; cancel deletes them.
- **No console flashes on Windows**: every subprocess uses `CREATE_NO_WINDOW`.
- **Error codes, not strings.** The backend returns `{"code":"ERR_LOGIN_REQUIRED","detail":"…"}`. The frontend maps each code to a localized title, a hint and a one-click fix (open cookie settings, network settings, install/update engine). Codes: `ERR_INVALID_URL, ERR_UNSUPPORTED_URL, ERR_LOGIN_REQUIRED, ERR_AGE_RESTRICTED, ERR_PRIVATE, ERR_GEO_BLOCKED, ERR_NOT_FOUND, ERR_RATE_LIMITED, ERR_DRM, ERR_NETWORK, ERR_TIMEOUT, ERR_PROXY, ERR_COOKIES, ERR_FFMPEG_MISSING, ERR_ENGINE_MISSING, ERR_ENGINE_INSTALL, ERR_CHECKSUM, ERR_DISK, ERR_CANCELED, ERR_LIVE_NOT_STARTED, ERR_FORMAT_UNAVAILABLE, ERR_SPOTIFY_UNSUPPORTED, ERR_SPOTIFY_MATCH, ERR_BUSY, ERR_JOB_NOT_FOUND, ERR_UNKNOWN`.
- **Engine sources**: yt-dlp from GitHub releases, ffmpeg + ffprobe from BtbN (Windows/Linux) or Martin Riedl (macOS), Deno from GitHub. yt-dlp uses Deno to list YouTube formats.
- **Diagnostics.** *Settings → Advanced → Copy diagnostic info* copies versions, tool paths, queue stats, settings (passwords and cookie paths redacted) and the last 150 log lines.

## Troubleshooting

**Behind a restricted network / firewall.** In *Settings → Network*, enable the proxy, choose HTTP/HTTPS/SOCKS5 and enter host and port (or paste a full URL like `socks5://user:pass@127.0.0.1:1080`). The proxy is used for analysis, downloads, Spotify metadata **and** engine installation/updates. If GitHub is unreachable even through the proxy, install the tools manually (next item).

**Engine won't install / offline machine.** Put the binaries in the data folder's `bin/` directory (`yt-dlp[.exe]`, `ffmpeg[.exe]`, `ffprobe[.exe]`, optionally `deno[.exe]`), or point *Settings → Download engine → Custom paths* at them. Anything on `PATH` (including Homebrew's `/opt/homebrew/bin`) is detected too.

**"ffmpeg is missing".** Click *Install engine* in the error card or in Settings. Without ffmpeg, separate video/audio streams can't be merged and audio can't be converted.

**"You need to be signed in" (Instagram, private/age-restricted/members-only videos, YouTube's "confirm you're not a bot").** In *Settings → Login & cookies*, choose *Use cookies from my browser* and pick the browser where you're logged in. On Windows, Chrome/Edge lock their cookie database while running: fully close the browser (check the tray) and retry, or export a `cookies.txt` (Netscape format) with a browser extension and use *Use a cookies.txt file*. Cookies grant account access, so keep that file private.

**"Too many requests".** Wait a few minutes, sign in with cookies, or route through a proxy.

**A site suddenly fails.** Sites change often. *Settings → Download engine → Update engine* gets the newest yt-dlp (auto-update is on by default).

**Spotify track matched the wrong version.** Matching uses `artist - title` on YouTube Music, with a YouTube search as fallback. Re-download from the original YouTube link if you need a specific version.

**Linux: blank window or build error about webkit2gtk-4.0.** On Ubuntu 24.04+ build with `-tags webkit2_41` (the build script does this automatically).

**No translucency.** Mica needs Windows 11 (build 22000+), and vibrancy needs macOS. Everything else uses a solid background. The toggle applies after restart.

## Design decisions

- **Wails over Electron**: a small native binary using the OS webview, Go for the process/queue work, and no bundled Chromium.
- **yt-dlp as the engine**: it supports more than 1,000 sites, is maintained daily and handles formats, cookies, merging and post-processing. Dlyven adds the UX, queue, i18n, error mapping and safety around it. Analysis uses `--flat-playlist`, so a 1,000-item playlist loads in seconds.
- **Engine binaries are fetched, not bundled**: smaller installs, licensing stays clean (ffmpeg GPL builds are fetched from their publishers), and yt-dlp can update independently of app releases.
- **Error codes over strings**: the UI can localize every failure and attach a fix, and the backend's classifier is unit-tested against real yt-dlp messages.
- **The frontend talks to `window.go.main.App` through a hand-typed bridge** (`lib/api.ts`) instead of importing generated bindings. The UI also runs in a plain browser with a mock backend, which makes design iteration and screenshots trivial.
- **Partial files next to the destination** (`.dlyven-partial/`), not in the app data dir, so finishing a 4 GB download is a rename instead of a cross-drive copy.
- **Visual language**: dark-first neutral surfaces (`#0B0D12`), 1px hairlines, soft layered shadows, one accent gradient for primary actions, an 8-pt spacing grid and 10–20 px radii. Accent fills and text pairs were picked to meet WCAG AA in both themes (amber uses dark text). In-app badges stay generic glyphs with a brand tint. README/site may show the platforms’ own marks next to a “not affiliated” line.
- **RTL is structural, not a patch**: logical properties (`ps-`, `me-`, `start-`, `end-`) everywhere, the sidebar and titlebar controls flip, directional icons mirror (media-playback icons intentionally don't), progress fills from the inline-start edge, and slide animations reverse. User and content text uses `dir="auto"`/`<bdi>`. Values with Latin units inside RTL sentences are wrapped in Unicode isolates so "393 MB of 610 MB" never scrambles. Dates fall back to neutral numeric format where the webview's ICU lacks a locale (Chromium has no `ckb` data).
- **Fonts are bundled** (Inter, Vazirmatn, Noto Sans Arabic, Noto Naskh Arabic via `@fontsource`): no CDN, so it works offline and on filtered networks. Inter is first in every stack; its unicode-range covers only Latin/Cyrillic/Greek, so English titles inside Persian/Kurdish UI keep Latin typography while Arabic-script text falls through to Vazirmatn/Noto. `npm run fonts:check` verifies every Sorani letter (ڕ ڵ ێ ۆ ە ڤ) and both digit sets are present.
- **Numbers & dates** always go through `Intl` with the active locale. There is an optional native-digits toggle (۰۱۲ for Persian, ٠١٢ for Kurdish) and a Jalali calendar option for Persian (Gregorian is forced explicitly, since `fa-IR` defaults to Persian in ICU).

## Adding a language

1. Copy `frontend/src/locales/en.json` to `frontend/src/locales/<code>.json` and translate the values (keep `{{placeholders}}`; add plural forms like `_few`/`_many` if the language needs them).
2. Register it in `frontend/src/i18n/index.ts`:

   ```ts
   import de from '../locales/de.json'
   // …
   { code: 'de', nativeName: 'Deutsch', dir: 'ltr', intlLocale: 'de-DE', resources: de },
   ```

   For an RTL language set `dir: 'rtl'`. Add `nativeNumbering` if it has its own digits.
3. Run `npm run i18n:check`. It fails on missing or extra keys and on placeholder mismatches.
4. (Optional) For a new script, add a font stack in `src/index.css` (`html[lang='xx'] { --font-sans: … }`) and a bundled font package.

## Adding a platform

Platforms are defined once, in Go:

1. Append to `Registry` in `internal/platform/platform.go`:

   ```go
   {ID: "kick", Name: "Kick", Color: "#53FC18", Icon: "play", Hosts: []string{"kick.com"}},
   ```

   `Icon` is any generic icon name mapped in `frontend/src/components/Platform.tsx`. Set `AudioOnly` to hide video tabs and `Matched` for metadata-matched services like Spotify.
2. Add a case to `TestDetect` in `internal/platform/platform_test.go`.

That's all. The frontend receives the registry from `App.GetAppInfo()` for live badges, filters and per-platform folders. Downloads work for anything yt-dlp supports even without a registry entry (it's shown as "Other site"). Platform-specific handling (like Spotify) lives in `internal/platform/<name>.go` and is wired in `extractor.Analyze` and the downloader's input selection.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Security reports: [SECURITY.md](SECURITY.md).

## License

Dlyven is licensed under the [GNU General Public License v3.0](LICENSE). Copyright © 2026 Amirreza Azadmehr.

ffmpeg/ffprobe are **not** shipped in this repository. On first launch the app may download GPL ffmpeg builds from their publishers; see [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

## Credits

- [yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [Deno](https://deno.com), [Wails](https://wails.io)
- Author: [Amirreza Azadmehr](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

## Assumptions & limitations

- **Spotify**: only public metadata is read (from Spotify's embed pages). Audio always comes from a matched public source; podcasts/shows aren't supported. If Spotify changes its embed page format, analysis returns `ERR_SPOTIFY_MATCH`.
- **Live streams** are rejected with a clear message rather than recorded.
- **Playlists** download each item with a quality *cap* (formats differ per item), not a fixed format ID.
- **Cookies from Chromium-based browsers on Windows** can require closing the browser (OS-level file lock). This is a yt-dlp/browser limitation.
- **System notifications** use the OS's own tools (PowerShell toast, `osascript`, `notify-send`). In-app toasts and the chime always work.
- **Translucency** is limited to Windows 11 and macOS; Linux uses the solid theme.
