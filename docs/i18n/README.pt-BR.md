<div dir="ltr" lang="pt-BR">

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

Cola o link, escolhe a qualidade, baixa. É só isso.

Desktop, código aberto, Windows / macOS / Linux. Por baixo é [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — o Dlyven é só uma janela mais bonita em volta.

**Site:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[README de desenvolvimento](../../README.md)**

**Idiomas:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">Esses logos não são nossos. YouTube, Instagram, TikTok, Spotify são deles — a gente só baixa.</p>

> Baixa só o que é seu ou o que você tem permissão de baixar.

## Instalar

1. Abre a [última versão](https://github.com/amirreza-azadmehr/dlyven/releases/latest).
2. Baixa **um** arquivo:

| Seu PC | Arquivo |
| --- | --- |
| Windows | `Dlyven-windows-amd64.exe` |
| Mac (Intel ou Apple silicon) | `Dlyven-macos-universal.zip` |
| Linux x64 | `Dlyven-linux-amd64` |
| Linux ARM | `Dlyven-linux-arm64` |

3. Abre. Windows: dá dois cliques no `.exe`. Linux: `chmod +x Dlyven-linux-*` e roda. Mac: descompacta e abre `Dlyven.app`.
4. Na primeira vez precisa de internet — ele mesmo puxa yt-dlp e ffmpeg. Depois é colar o link e baixar.

Sem instalador. Os builds oficiais são assinados quando os certificados da Apple / Windows estão no repo. Do código? [README de desenvolvimento](../../README.md).

## Como usar

1. Cola um link (ou <kbd>Ctrl/⌘ V</kbd>, ou solta uma URL / um `.txt` de links).
2. Um segundo — título, thumbnail, todas as qualidades que achar.
3. Vídeo, áudio, legendas ou thumbnail. Tem preset se você não quiser pensar.
4. Escolhe a pasta, baixa. Pausa / continua / cancela na fila.

Post privado: Configurações → Login e cookies. Rede travada: Configurações → Rede.

## Sites

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… e [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Link desconhecido também baixa; aparece como “Outro site”.

**YouTube** — cola o link, vídeo ou MP3, baixa. Playlist tem checklist.

**Instagram** — URL do reel ou post. Privado precisa estar logado (cookies).

**TikTok** — cola, vídeo ou áudio. Link curto vale.

**Spotify** — a gente não puxa o áudio do Spotify. Lê título/artista público, acha um match público e etiqueta. Sem podcast.

## FAQ

**App oficial do YouTube / Instagram / TikTok?** Não. Só uma UI de desktop em cima do yt-dlp.

**Baixa música direto do Spotify?** Não. Só casa com uma gravação pública.

**Manda telemetria?** Não.

**Cadê meus arquivos?** Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Contribuir

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## Licença

[GPL-3.0](../../LICENSE) © 2026 Amirreza Azadmehr. [Avisos de terceiros](../../THIRD_PARTY_NOTICES.md).

Autor: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
