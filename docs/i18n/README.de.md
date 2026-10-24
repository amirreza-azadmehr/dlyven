<div dir="ltr" lang="de">

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

Link rein, Qualität wählen, Download. Mehr braucht’s nicht.

Desktop, Open Source, Windows / macOS / Linux. Darunter [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — Dlyven ist nur ein netteres Fenster drumherum.

**Seite:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[Entwickler-README](../../README.md)**

**Sprachen:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">Die Logos sind nicht unsere. YouTube, Instagram, TikTok, Spotify gehören denen — wir sind nur ein Downloader.</p>

> Lad nur runter, was dir gehört oder wofür du die Erlaubnis hast.

## Installation

Lad **eine** Datei runter (die Links zeigen immer auf den neuesten Release):

| Dein Rechner | Download |
| --- | --- |
| Windows 10 / 11, Installer (empfohlen) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| Windows 10 / 11, portabel (ohne Installation) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| Mac (Apple Silicon oder Intel) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| Linux x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| Linux ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

Dann: Windows: Setup ausführen oder die portable `.exe` doppelklicken. Mac: `.dmg` öffnen und Dlyven auf Applications ziehen. Linux: `chmod +x Dlyven-Linux-*` und starten (braucht WebKitGTK 4.1: Ubuntu 22.04+ / Debian 12+).

Beim ersten Start braucht’s Internet — yt-dlp und ffmpeg holt er selbst. Danach Link rein, Download.

**Updates:** Gibt’s eine neue Version, sagt Dlyven Bescheid. **Settings → Updates → Update now**. Keine Neuinstallation.

Sicherheitswarnung beim ersten Start? [Hier steht, was zu tun ist](../../README.md#installation-and-security-warnings). Aus dem Quellcode? [Entwickler-README](../../README.md).

## So geht’s

1. Link einfügen (oder <kbd>Ctrl/⌘ V</kbd>, oder URL / `.txt` mit Links draufziehen).
2. Kurz warten — Titel, Vorschaubild, alle Qualitäten, die es findet.
3. Video, Audio, Untertitel oder Thumbnail. Presets, wenn du nicht nachdenken willst.
4. Ordner wählen, Download. Pause / weiter / abbrechen in der Queue.

Private Posts: Einstellungen → Login & Cookies. Eingeschränktes Netz: Einstellungen → Netzwerk.

## Seiten

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… und [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Unbekannte Links gehen auch; da steht dann „Andere Seite“.

**YouTube** — Link rein, Video oder MP3, Download. Playlists haben eine Checkliste.

**Instagram** — Reel- oder Post-URL. Privates braucht Login (Cookies).

**TikTok** — rein, Video oder Audio. Kurzlinks gehen.

**Spotify** — Audio kommt nicht von Spotify. Wir lesen Titel/Artist öffentlich, suchen einen öffentlichen Match, taggen die Datei. Keine Podcasts.

## FAQ

**Offizielle YouTube- / Instagram- / TikTok-App?** Nö. Nur eine Desktop-UI über yt-dlp.

**Songs direkt von Spotify?** Nein. Wir matchen nur eine öffentliche Aufnahme.

**Phone home?** Keine Telemetrie. Er spricht nur mit GitHub, um seine Tools zu holen und nach Updates zu schauen (die automatische Prüfung lässt sich in den Settings abschalten).

**Wo liegt meine Data?** Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Mitmachen

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## Lizenz

[GPL-3.0](../../LICENSE) © 2026 Amirreza Azadmehr. [Hinweise Dritter](../../THIRD_PARTY_NOTICES.md).

Autor: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
