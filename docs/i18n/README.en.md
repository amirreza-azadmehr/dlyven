<div dir="ltr" lang="en">

# Dlyven

<p align="center">
  <img src="../images/home.png" alt="Home" width="880">
</p>
<p align="center">
  <img src="../images/formats.png" alt="Formats" width="880">
</p>
<p align="center">
  <img src="../images/queue.png" alt="Queue" width="880">
</p>

Paste a link, pick a quality, hit download. That’s the whole app.

Desktop, open source, Windows / macOS / Linux. Under the hood it’s [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — Dlyven is just a nicer window around them.

**Site:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[Full developer README](../../README.md)**

**Languages:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

<p align="center">
  <img src="../images/brands/youtube.svg" alt="YouTube" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/instagram.svg" alt="Instagram" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/tiktok.svg" alt="TikTok" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/spotify.svg" alt="Spotify" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/x.svg" alt="X" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/facebook.svg" alt="Facebook" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/soundcloud.svg" alt="SoundCloud" height="28">
  &nbsp;&nbsp;
  <img src="../images/brands/twitch.svg" alt="Twitch" height="28">
</p>

<p align="center">Those logos aren’t ours. YouTube, Instagram, TikTok, Spotify and the rest belong to them — we’re just a downloader.</p>

> Only grab stuff you own or have permission to download.

## Install

Download **one** file (the links always point to the newest release):

| Your computer | Download |
| --- | --- |
| Windows 10 / 11, installer (recommended) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| Windows 10 / 11, portable (no install) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| Mac (Apple silicon or Intel) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| Linux x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| Linux ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

Then: Windows: run the Setup, or double-click the portable `.exe`. Mac: open the `.dmg` and drag Dlyven onto Applications. Linux: `chmod +x Dlyven-Linux-*` then run it (needs WebKitGTK 4.1: Ubuntu 22.04+ / Debian 12+).

First launch needs internet — it grabs yt-dlp and ffmpeg by itself. After that, paste a link and download.

**Updates:** when a new version is out, Dlyven tells you. **Settings → Updates → Update now**. No reinstall.

Security warning on first launch? [Here’s what to do](../../README.md#installation-and-security-warnings). Building from source? [developer README](../../README.md).

## How to use

1. Paste a link (or hit <kbd>Ctrl/⌘ V</kbd>, or drop a URL / a `.txt` of links).
2. Wait a second — title, thumbnail, every quality it can find.
3. Video, audio, subtitles or thumbnail. Presets if you don’t want to think.
4. Pick a folder, download. Pause / resume / cancel from the queue.

Private posts: Settings → Login & cookies. Restricted network: Settings → Network.

## Sites

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… and [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Unknown links still work; they just show up as “Other site”.

**YouTube** — paste the link, pick video or MP3, download. Playlists get a checklist.

**Instagram** — reel or post URL. Private ones need you to be logged in (cookies).

**TikTok** — paste it, video or audio. Short links are fine.

**Spotify** — we don’t pull audio from Spotify. We read the public title/artist, find a public match, and tag the file. Podcasts aren’t a thing here.

## FAQ

**Official YouTube / Instagram / TikTok app?** Nope. Just a desktop UI on top of yt-dlp.

**Rip songs straight from Spotify?** No. We only match a public recording and tag it.

**Does it phone home?** No telemetry. It only talks to GitHub to fetch its tools and check for updates (the automatic checks can be turned off in Settings).

**Where’s my stuff?** Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Contributing

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## License

[GPL-3.0](../../LICENSE) © 2026 Amirreza Azadmehr. [Third-party notices](../../THIRD_PARTY_NOTICES.md).

Author: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
