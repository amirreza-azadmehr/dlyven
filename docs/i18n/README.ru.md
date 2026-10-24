<div dir="ltr" lang="ru">

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

Вставь ссылку, выбери качество, скачай. Вот и всё.

Десктоп, опенсорс, Windows / macOS / Linux. Под капотом [yt-dlp](https://github.com/yt-dlp/yt-dlp) + ffmpeg — Dlyven просто нормальное окно вокруг них.

**Сайт:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[README для разработчиков](../../README.md)**

**Языки:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">Это не наши логотипы. YouTube, Instagram, TikTok, Spotify — их. Мы просто качалка.</p>

> Качай только своё или то, на что есть разрешение.

## Установка

Скачай **один** файл (ссылки всегда ведут на свежий релиз):

| Твой комп | Скачать |
| --- | --- |
| Windows 10 / 11, установщик (рекомендуется) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| Windows 10 / 11, портативная (без установки) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| Mac (Apple silicon или Intel) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| Linux x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| Linux ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

Дальше: Windows: запусти Setup или дважды кликни по портативному `.exe`. Mac: открой `.dmg` и перетащи Dlyven в Applications. Linux: `chmod +x Dlyven-Linux-*` и запускай (нужен WebKitGTK 4.1: Ubuntu 22.04+ / Debian 12+).

Первый раз нужен интернет — yt-dlp и ffmpeg подтянутся сами. Дальше вставляешь ссылку и качаешь.

**Обновления:** когда выходит новая версия, Dlyven сам скажет. **Settings → Updates → Update now**. Переустанавливать не надо.

Предупреждение безопасности при первом запуске? [Смотри тут](../../README.md#installation-and-security-warnings). Из исходников? [README для разработчиков](../../README.md).

## Как пользоваться

1. Вставь ссылку (или <kbd>Ctrl/⌘ V</kbd>, или кинь URL / `.txt` со ссылками).
2. Секунда — название, превью, все качества, что найдёт.
3. Видео, аудио, субтитры или превью. Есть пресеты, если лень выбирать.
4. Папка, скачать. Пауза / продолжить / отмена в очереди.

Закрытые посты: Настройки → Вход и cookies. Сеть с фильтром: Настройки → Сеть.

## Сайты

YouTube, Instagram, TikTok, Spotify, X, Facebook, Vimeo, SoundCloud, Reddit, Twitch… и [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). Незнакомые ссылки тоже качаются — пишется «Другой сайт».

**YouTube** — вставь ссылку, видео или MP3, скачай. У плейлистов чеклист.

**Instagram** — URL рилса или поста. Закрытое — надо быть залогиненным (cookies).

**TikTok** — вставь, видео или аудио. Короткие ссылки ок.

**Spotify** — аудио оттуда не берём. Читаем публичные название/артиста, ищем публичную запись, тегируем. Подкастов нет.

## FAQ

**Официальное приложение YouTube / Instagram / TikTok?** Нет. Просто десктопный UI поверх yt-dlp.

**Песни напрямую со Spotify?** Нет. Только матчим публичную запись.

**Шлёт куда-то статистику?** Нет. Только ходит на GitHub за своими инструментами и проверкой обновлений (автопроверку можно выключить в Settings).

**Где мои файлы?** Windows `%AppData%\Dlyven` · Mac `~/Library/Application Support/Dlyven` · Linux `~/.config/Dlyven`.

## Участие

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## Лицензия

[GPL-3.0](../../LICENSE) © 2026 Amirreza Azadmehr. [Сторонние уведомления](../../THIRD_PARTY_NOTICES.md).

Автор: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
