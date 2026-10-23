# Third-party notices

Dlyven is licensed under the [GNU GPL-3.0](LICENSE). This file lists other works
that Dlyven uses, with the licenses that require attribution.

ffmpeg/ffprobe, yt-dlp, and Deno are **not** stored in this repository. The
running app may download them on first launch from their publishers, then verify
SHA-256 checksums. See `internal/engine/sources.go`.

## Runtime engines (downloaded, not vendored)

| Component | Source | License | Notes |
| --- | --- | --- | --- |
| [yt-dlp](https://github.com/yt-dlp/yt-dlp) | GitHub Releases | Unlicense | Extractor / downloader |
| ffmpeg / ffprobe | [BtbN/FFmpeg-Builds](https://github.com/BtbN/FFmpeg-Builds) (Windows, Linux, **GPL** static) and [Martin Riedl](https://ffmpeg.martin-riedl.de/) (macOS) | GPL (upstream FFmpeg also has LGPL builds; Dlyven fetches the **GPL** builds) | Merge, convert, thumbnails |
| [Deno](https://github.com/denoland/deno) | GitHub Releases | MIT | JS runtime yt-dlp uses for some YouTube formats |

If you redistribute a Dlyven build **together with** those binaries, you must
also meet *their* license terms (especially ffmpeg GPL: corresponding source).

## Direct Go modules

| Module | License |
| --- | --- |
| github.com/wailsapp/wails/v2 | MIT |
| github.com/ulikunitz/xz | BSD-3-Clause |
| golang.org/x/sys | BSD-3-Clause |
| golang.org/x/text | BSD-3-Clause |
| modernc.org/sqlite (and modernc.org/libc, mathutil, memory) | BSD-3-Clause |

Wails pulls in further MIT/BSD libraries (Echo, WebView2 bindings, and others).
See `go.mod` / `go.sum`. None of those are copyleft in a way that conflicts with
GPL-3.0 for this application.

## Direct npm packages

| Package | License |
| --- | --- |
| react, react-dom, react-i18next, i18next | MIT |
| framer-motion, zustand, clsx | MIT |
| @radix-ui/* | MIT |
| lucide-react | ISC |
| vite, typescript, tailwindcss, postcss, autoprefixer, @vitejs/plugin-react | MIT |

## Fonts (bundled in the app and on the website)

| Family | Package | License |
| --- | --- | --- |
| Inter | @fontsource-variable/inter | SIL Open Font License 1.1 |
| Vazirmatn | @fontsource-variable/vazirmatn | SIL Open Font License 1.1 |
| Noto Sans Arabic | @fontsource-variable/noto-sans-arabic | SIL Open Font License 1.1 |
| Noto Naskh Arabic | @fontsource-variable/noto-naskh-arabic | SIL Open Font License 1.1 |

OFL fonts may be used and redistributed if the license is kept and the reserved
font names are not used for derivatives without permission.

## Icons

In-app icons are [Lucide](https://lucide.dev) (ISC). Platform badges use generic
glyphs plus a tint; **no** third-party brand logos are shipped.

## Compatibility with GPL-3.0

MIT, BSD, ISC, Unlicense, and SIL OFL may be combined with a GPL-3.0
application. The GPL ffmpeg builds downloaded at runtime are themselves GPL;
keeping them out of this git tree avoids shipping ffmpeg sources here, but a
binary distribution that includes ffmpeg must follow ffmpeg’s GPL.
