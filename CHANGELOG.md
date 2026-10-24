# Changelog

The real notes are on [GitHub Releases](https://github.com/amirreza-azadmehr/dlyven/releases). Push app code to `main` and GitHub cuts the next patch, builds it, and writes the notes from the commits. Docs/site-only pushes don’t.

## [Unreleased]

Whatever’s on `main` that isn’t tagged yet.

- **Updates itself.** Settings → Updates checks GitHub on launch, then downloads the release for your system, verifies it against `SHA256SUMS.txt` and installs it on restart. Windows Setup installs run the new Setup silently; portable, macOS and Linux copies swap in place. Turn the startup check off in the same place.
- **Release files say what they’re for:** `Dlyven-Windows-x64-Setup.exe`, `Dlyven-Windows-x64-Portable.exe`, `Dlyven-macOS-Universal.dmg` / `.zip`, `Dlyven-Linux-x64`, `Dlyven-Linux-ARM64`. Names no longer change between releases, so `releases/latest/download/<file>` links keep working. Each release opens with a download table.
- **The macOS `.dmg` opens as a drag-to-Applications window** (Dlyven, an arrow, Applications).
- README: direct download links, Linux runtime requirements (WebKitGTK 4.1, glibc 2.34+), an Updating section, and Wails v2.12.0 to match CI.

## [1.0.1] — 2026-10-24

Home opens immediately instead of sitting on a blank shell. Search no longer waits for `yt-dlp --version` (that hung ~45s on macOS and made Instagram look like a one-minute load). Version probes run in the background with a 3s cap; Gatekeeper quarantine is cleared on managed binaries.

## [1.0.0] — 2026-10-23

First public Dlyven. Used to be a private app called GrabIt.

[Unreleased]: https://github.com/amirreza-azadmehr/dlyven/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/amirreza-azadmehr/dlyven/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/amirreza-azadmehr/dlyven/releases/tag/v1.0.0
