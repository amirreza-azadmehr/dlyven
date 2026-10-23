// Verifies that the bundled Arabic-script fonts contain every Sorani/Persian glyph.
// Usage: npm run fonts:check
// Minimal WOFF2 cmap reader: decompress the Brotli stream, find the (untransformed) cmap
// table and read format 4 / 12 subtables.
import fs from 'node:fs'; import zlib from 'node:zlib';
const KNOWN = ['cmap','head','hhea','hmtx','maxp','name','OS/2','post','cvt ','fpgm','glyf','loca','prep','CFF ','VORG','EBDT','EBLC','gasp','hdmx','kern','LTSH','PCLT','VDMX','vhea','vmtx','BASE','GDEF','GPOS','GSUB','EBSC','JSTF','MATH','CBDT','CBLC','COLR','CPAL','SVG ','sbix','acnt','avar','bdat','bloc','bsln','cvar','fdsc','feat','fmtx','fvar','gvar','hsty','just','lcar','mort','morx','opbd','prop','trak','Zapf','Silf','Glat','Gloc','Feat','Sill'];
function b128(b, o) { let v = 0; for (let i = 0; i < 5; i++) { const x = b[o.p++]; v = (v << 7) | (x & 0x7f); if (!(x & 0x80)) return v; } }
function cmapOf(path) {
  const b = fs.readFileSync(path); const dv = new DataView(b.buffer, b.byteOffset);
  const numTables = dv.getUint16(12); const totalCompressed = dv.getUint32(20);
  const o = { p: 48 }; const tables = []; let off = 0;
  for (let i = 0; i < numTables; i++) {
    const flags = b[o.p++]; const idx = flags & 63;
    const tag = idx === 63 ? b.subarray(o.p, (o.p += 4)).toString('latin1') : KNOWN[idx];
    const xform = (flags >> 6) & 3; const orig = b128(b, o);
    let len = orig; if ((tag === 'glyf' || tag === 'loca') ? xform === 0 : xform !== 0) len = b128(b, o);
    tables.push({ tag, off, len }); off += len;
  }
  const data = zlib.brotliDecompressSync(b.subarray(o.p, o.p + totalCompressed));
  const t = tables.find((x) => x.tag === 'cmap'); const c = new DataView(data.buffer, data.byteOffset + t.off, t.len);
  const set = new Set(); const n = c.getUint16(2);
  for (let i = 0; i < n; i++) {
    const so = c.getUint32(4 + i * 8 + 4); const fmt = c.getUint16(so);
    if (fmt === 4) {
      const segX2 = c.getUint16(so + 6); const ends = so + 14, starts = ends + segX2 + 2, deltas = starts + segX2, ros = deltas + segX2;
      for (let s = 0; s < segX2 / 2; s++) {
        const end = c.getUint16(ends + 2 * s), start = c.getUint16(starts + 2 * s), delta = c.getInt16(deltas + 2 * s), ro = c.getUint16(ros + 2 * s);
        for (let ch = start; ch <= end && ch !== 0xffff; ch++) {
          let g = ro === 0 ? (ch + delta) & 0xffff : c.getUint16(ros + 2 * s + ro + 2 * (ch - start));
          if (ro !== 0 && g !== 0) g = (g + delta) & 0xffff; if (g) set.add(ch);
        }
      }
    } else if (fmt === 12) {
      const groups = c.getUint32(so + 12);
      for (let g = 0; g < groups; g++) { const a = c.getUint32(so + 16 + g * 12), z = c.getUint32(so + 20 + g * 12); for (let ch = a; ch <= z; ch++) set.add(ch); }
    }
  }
  return set;
}
const sets = { sorani: 'ئابپتجچحخدرڕزژسشعغفڤقکگلڵمنهەوۆیێ‌', persian: 'آابپتثجچحخدذرزژسشصضطظعغفقکگلمنوهی۰۱۲۳۴۵۶۷۸۹', 'arabic-indic digits': '٠١٢٣٤٥٦٧٨٩' };
const files = process.argv.length > 2 ? process.argv.slice(2) : [
  "node_modules/@fontsource-variable/noto-sans-arabic/files/noto-sans-arabic-arabic-wght-normal.woff2",
  "node_modules/@fontsource-variable/noto-naskh-arabic/files/noto-naskh-arabic-arabic-wght-normal.woff2",
  "node_modules/@fontsource-variable/vazirmatn/files/vazirmatn-arabic-wght-normal.woff2",
];
let failed = false;
for (const f of files) { const cm = cmapOf(f);
  for (const [k, s] of Object.entries(sets)) { const miss = [...s].filter((ch) => !cm.has(ch.codePointAt(0))).map((ch) => `${ch} U+${ch.codePointAt(0).toString(16).toUpperCase()}`);
    console.log(f.split('/').pop().padEnd(48), k.padEnd(20), miss.length ? 'MISSING ' + miss.join(" ") : `OK (${cm.size} codepoints)`); if (miss.length) failed = true; } }
process.exit(failed ? 1 : 0);
