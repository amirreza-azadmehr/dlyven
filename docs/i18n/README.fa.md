<div dir="rtl" lang="fa">

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

لینک را وارد کن، کیفیت را انتخاب کن، دانلود. تمام.

برنامهٔ دسکتاپ متن‌باز برای ویندوز، مک و لینوکس. کار اصلی را [yt-dlp](https://github.com/yt-dlp/yt-dlp) و ffmpeg می‌کنند؛ دلیون ظاهر ساده‌تری برای همان‌هاست.

**سایت:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[README کامل برای برنامه‌نویس](../../README.md)**

**زبان‌ها:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">این لوگوها مال ما نیستن. یوتیوب، اینستا، تیک‌تاک، اسپاتیفای مال خودشن — ما فقط یه دانلودریم.</p>

> فقط چیزی رو دانلود کن که مال خودته یا اجازه‌ش رو داری.

## نصب

**یه فایل** بردار، همونی که مال سیستمته (لینک‌ها همیشه آخرین نسخه رو می‌دن):

| سیستمت | دانلود |
| --- | --- |
| ویندوز 10 / 11، نصب‌کننده (پیشنهادی) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| ویندوز 10 / 11، پرتابل (بدون نصب) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| مک (اپل سیلیکون یا اینتل) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| لینوکس x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| لینوکس ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

بعدش: ویندوز: Setup رو اجرا کن، یا روی `.exe` پرتابل دابل‌کلیک کن. مک: `.dmg` رو باز کن و Dlyven رو بکش روی Applications. لینوکس: اول `chmod +x Dlyven-Linux-*` بعد اجرا (WebKitGTK 4.1 لازمه: اوبونتو 22.04+ / دبیان 12+).

بار اول اینترنت می‌خواد — خودش yt-dlp و ffmpeg رو می‌گیره. بعدش لینک رو بچسبون و دانلود کن.

**آپدیت:** نسخهٔ جدید که بیاد، خود Dlyven خبرت می‌کنه. **تنظیمات ← به‌روزرسانی ← به‌روزرسانی**. نصب دوباره لازم نیست.

بار اول هشدار امنیتی دیدی؟ [اینجا رو ببین](../../README.md#installation-and-security-warnings). از سورس؟ [README برنامه‌نویس](../../README.md).

## چطور کار می‌کنه

1. لینک رو بچسبون (یا <kbd>Ctrl/⌘ V</kbd>، یا یه URL / فایل `.txt` پر از لینک بنداز روش).
2. یه کم صبر کن — عنوان، عکس، همه‌چیز کیفیت‌هایی که پیدا می‌کنه.
3. ویدیو، صدا، زیرنویس یا تامبنیل. اگه حوصله نداری از پیش‌فرض‌ها استفاده کن.
4. پوشه رو بزن، دانلود. تو صف می‌تونی مکث / ادامه / لغو کنی.

پست خصوصی: تنظیمات → ورود و کوکی. نت محدود: تنظیمات → شبکه.

## سایت‌ها

یوتیوب، اینستا، تیک‌تاک، اسپاتیفای، ایکس، فیسبوک، ویمو، ساندکلاد، ردیت، توییچ… و [whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). لینک ناشناس هم کار می‌کنه؛ می‌نویسه «سایت دیگر».

**یوتیوب** — لینک رو بچسبون، ویدیو یا MP3، دانلود. پلی‌لیست چک‌لیست داره.

**اینستاگرام** — لینک ریل یا پست. خصوصی‌ها باید لاگین باشی (کوکی).

**تیک‌تاک** — بچسبون، ویدیو یا صدا. لینک کوتاه هم اوکیه.

**اسپاتیفای** — صدا رو از خود اسپاتیفای نمی‌کشیم. اسم آهنگ رو می‌خونیم، یه نسخهٔ عمومی پیدا می‌کنیم، تگ می‌زنیم. پادکست نداریم.

## سؤال‌های پرتکرار

**اپ رسمی یوتیوب / اینستا / تیک‌تاکه؟** نه بابا. فقط یه ظاهر رومیزی روی yt-dlp.

**آهنگ رو مستقیم از اسپاتیفای می‌کشه؟** نه. فقط یه نسخهٔ عمومی رو مچ می‌کنه.

**جایی گزارش می‌فرسته؟** نه. هیچ آماری نمی‌فرسته؛ فقط برای گرفتن ابزارهاش و چک کردن آپدیت به گیت‌هاب وصل می‌شه (چک خودکار رو توی تنظیمات می‌شه خاموش کرد).

**فایلام کجاست؟** ویندوز `%AppData%\Dlyven` · مک `~/Library/Application Support/Dlyven` · لینوکس `~/.config/Dlyven`.

## مشارکت

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## مجوز

[GPL-3.0](../../LICENSE) © ۲۰۲۶ امیررضا آزادمهر. [اطلاعیه‌های شخص ثالث](../../THIRD_PARTY_NOTICES.md).

سازنده: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
