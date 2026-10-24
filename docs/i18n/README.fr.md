<div dir="ltr" lang="fr">

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

Colle un lien, choisis la qualité, télécharge. C’est tout.

Bureau, open source, Windows / macOS / Linux. En dessous c’est [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — Dlyven, c’est juste une fenêtre plus sympa.

**Site:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[README développeur](../../README.md)**

**Langues:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">Ces logos ne sont pas à nous. YouTube, Instagram, TikTok, Spotify, tout ça c’est à eux — on est juste un téléchargeur.</p>

> Ne télécharge que ce qui t’appartient ou ce que tu as le droit de prendre.

## Installer

Télécharge **un** fichier (les liens pointent toujours vers la dernière version) :

| Ton ordi | Téléchargement |
| --- | --- |
| Windows 10 / 11, installeur (recommandé) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| Windows 10 / 11, portable (sans installation) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| Mac (Apple silicon ou Intel) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| Linux x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| Linux ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

Ensuite : Windows : lance le Setup, ou double-clique sur le `.exe` portable. Mac : ouvre le `.dmg` et glisse Dlyven sur Applications. Linux : `chmod +x Dlyven-Linux-*` puis lance (il faut WebKitGTK 4.1 : Ubuntu 22.04+ / Debian 12+).

La première fois il faut internet — il récupère yt-dlp et ffmpeg tout seul. Ensuite tu colles un lien et c’est parti.

**Mises à jour :** quand une nouvelle version sort, Dlyven te prévient. **Settings → Updates → Update now**. Pas besoin de réinstaller.

Un avertissement de sécurité au premier lancement ? [Voir ici](../../README.md#installation-and-security-warnings). Depuis le code ? [README développeur](../../README.md).

## Comment ça marche

1. Colle un lien (ou <kbd>Ctrl/⌘ V</kbd>, ou dépose une URL / un `.txt` de liens).
2. Une seconde — titre, miniature, toutes les qualités trouvées.
3. Vidéo, audio, sous-titres ou miniature. Des presets si tu n’as pas envie de réfléchir.
4. Choisis un dossier, télécharge. Pause / reprise / annulation dans la file.

Contenu privé : Réglages → Connexion et cookies. Réseau filtré : Réglages → Réseau.

## Sites

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… et [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Les liens inconnus marchent aussi ; ça affiche « Autre site ».

**YouTube** — colle le lien, vidéo ou MP3, télécharge. Les playlists ont une checklist.

**Instagram** — URL du reel ou du post. Le privé, il faut être connecté (cookies).

**TikTok** — colle, vidéo ou audio. Les liens courts passent.

**Spotify** — on ne prend pas l’audio chez Spotify. On lit le titre/artiste public, on trouve un match public, on tague. Pas de podcasts.

## FAQ

**App officielle YouTube / Instagram / TikTok ?** Non. Juste une UI bureau par-dessus yt-dlp.

**Je pique les morceaux chez Spotify ?** Non. On matche juste un enregistrement public.

**Ça phone home ?** Pas de télémétrie. Il parle seulement à GitHub pour récupérer ses outils et vérifier les mises à jour (la vérification automatique se coupe dans Settings).

**Mes fichiers ?** Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Contribuer

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## Licence

[GPL-3.0](../../LICENSE) © 2026 Amirreza Azadmehr. [Mentions tierces](../../THIRD_PARTY_NOTICES.md).

Auteur : [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
