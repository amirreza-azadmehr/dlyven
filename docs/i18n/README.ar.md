<div dir="rtl" lang="ar">

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

لزّق الرابط، اختَر الجودة، نزّل. هذا كل شيء.

سطح مكتب، مفتوح المصدر، ويندوز / ماك / لينكس. تحت الغطاء [yt-dlp](https://github.com/yt-dlp/yt-dlp) وffmpeg — دليفن بس نافذة ألطف.

**الموقع:** [amirreza-azadmehr.github.io/dlyven](https://amirreza-azadmehr.github.io/dlyven/) · **[README المطوّر](../../README.md)**

**اللغات:** [English](README.en.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [Français](README.fr.md) · [Türkçe](README.tr.md) · [العربية](README.ar.md) · [Русский](README.ru.md) · [简体中文](README.zh-CN.md) · [Deutsch](README.de.md) · [Português (Brasil)](README.pt-BR.md) · [کوردی](README.ckb.md)

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

<p align="center">هذي الشعارات مو لنا. يوتيوب وإنستغرام وتيك توك وسبوتيفاي حقهم — إحنا بس برنامج تنزيل.</p>

> نزّل بس اللي تملكه أو اللي مسموح لك تنزّله.

## التثبيت

نزّل **ملف واحد** (الروابط دايمًا لآخر إصدار):

| جهازك | التنزيل |
| --- | --- |
| ويندوز 10 / 11، مثبّت (مستحسن) | [Dlyven-Windows-x64-Setup.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Setup.exe) |
| ويندوز 10 / 11، نسخة محمولة (بدون تثبيت) | [Dlyven-Windows-x64-Portable.exe](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Windows-x64-Portable.exe) |
| ماك (Apple silicon أو إنتل) | [Dlyven-macOS-Universal.dmg](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-macOS-Universal.dmg) |
| لينكس x64 | [Dlyven-Linux-x64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-x64) |
| لينكس ARM64 | [Dlyven-Linux-ARM64](https://github.com/amirreza-azadmehr/dlyven/releases/latest/download/Dlyven-Linux-ARM64) |

بعدها: ويندوز: شغّل الـSetup، أو دبل كلك على الـ`.exe` المحمول. ماك: افتح الـ`.dmg` واسحب Dlyven على Applications. لينكس: `chmod +x Dlyven-Linux-*` بعدين شغّله (يحتاج WebKitGTK 4.1: أوبونتو 22.04+ / ديبيان 12+).

أول مرة يحتاج نت — يجيب yt-dlp وffmpeg لحاله. بعدين لزّق الرابط ونزّل.

**التحديثات:** لما ينزل إصدار جديد، Dlyven يخبرك. **Settings → Updates → Update now**. بدون إعادة تثبيت.

تحذير أمان أول تشغيل؟ [شوف هنا](../../README.md#installation-and-security-warnings). من السورس؟ [README المطوّر](../../README.md).

## طريقة الاستخدام

1. لزّق رابط (أو <kbd>Ctrl/⌘ V</kbd>، أو ارمي URL / ملف `.txt` مليان روابط).
2. ثانية — العنوان، الصورة، كل الجودات اللي يلقاها.
3. فيديو، صوت، ترجمة أو صورة. فيه اختصارات لو ما تبي تفكر.
4. اختَر مجلد، نزّل. من الطابور وقف / كمّل / ألغي.

منشور خاص: الإعدادات → الدخول والكوكيز. نت مقيّد: الإعدادات → الشبكة.

## المواقع

يوتيوب، إنستغرام، تيك توك، سبوتيفاي، X، فيسبوك، فيميو، ساوندكلاود، ردِت، تويتش… و[whatever else yt-dlp knows](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). الروابط الغريبة تشتغل بعد؛ يكتب «موقع آخر».

**يوتيوب** — لزّق الرابط، فيديو أو MP3، نزّل. قوائم التشغيل فيها قائمة اختيار.

**إنستغرام** — رابط الريل أو المنشور. الخاص يحتاج تكون مسجّل (كوكيز).

**تيك توك** — لزّقه، فيديو أو صوت. الروابط القصيرة تمام.

**سبوتيفاي** — ما نسحب الصوت من سبوتيفاي. نقرأ العنوان/الفنان العام، نلقى تسجيل عام، نعلّمه. ما فيه بودكاست.

## أسئلة

**تطبيق رسمي ليوتيوب / إنستغرام / تيك توك؟** لا. بس واجهة سطح مكتب فوق yt-dlp.

**أغاني مباشرة من سبوتيفاي؟** لا. بس نطابق تسجيل عام.

**يرسل تقارير؟** لا. ما فيه أي تتبّع؛ بس يتصل بـGitHub عشان يجيب أدواته ويشيك على التحديثات (تقدر توقف الفحص التلقائي من الإعدادات).

**وين بياناتي؟** ويندوز `%AppData%\Dlyven` · ماك `~/Library/Application Support/Dlyven` · لينكس `~/.config/Dlyven`.

## المساهمة

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) · [SECURITY.md](../../SECURITY.md)

## الرخصة

[GPL-3.0](../../LICENSE) © 2026 أميررضا آزادمهر. [إشعارات الطرف الثالث](../../THIRD_PARTY_NOTICES.md).

المؤلف: [GitHub](https://github.com/amirreza-azadmehr) · [LinkedIn](https://tr.linkedin.com/in/amirreza-azadmehr) · [Telegram](https://t.me/amir_reza_azadmehr)

</div>
