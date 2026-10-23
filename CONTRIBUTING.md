# Contributing to Dlyven

Thanks for helping. Dlyven is GPL-3.0: contributions are accepted under the same license.

## Before you start

1. Search [issues](https://github.com/amirreza-azadmehr/dlyven/issues) for a duplicate.
2. For a large change, open an issue first.
3. Keep the legal line: only download content you own or have permission to download. Do not add DRM circumvention or “bypass paywall” features.

## Development

```bash
cd frontend && npm install && cd ..
wails dev
```

UI-only (mock backend): `cd frontend && npm run dev`

Logs live in the app data folder. `DLYVEN_DEBUG=1` also prints them in the terminal.

### Tests

```bash
go test ./internal/...
cd frontend && npm run typecheck && npm run i18n:check && npm run fonts:check
```

`go test -tags smoke -v ./internal/smoke` hits the real internet; it is optional.

### Build

See [README.md](README.md#building) or `./scripts/build.sh`.

## Project map

| Area | Where |
| --- | --- |
| Product name / data dir | `internal/appinfo` |
| Platforms | `internal/platform` |
| yt-dlp / ffmpeg install | `internal/engine` |
| Queue | `internal/downloader` |
| UI strings | `frontend/src/locales/*.json` |
| Author links | `frontend/src/config/author.ts` |

Rename the app only through `appinfo` + `wails.json` + `frontend/src/config/app.ts`.

## Pull requests

- One concern per PR.
- Run the tests above.
- Keep locale keys in sync (`npm run i18n:check`).
- Do not commit `frontend/dist`, `frontend/wailsjs`, `build/bin`, `.env`, or cookies.
- New user-visible strings go in **all** locale files.

## Code of conduct

[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
