// Verifies every locale has exactly the keys of en.json, with the same {{placeholders}},
// and that no translation is empty. Runs before every build (npm run build).
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const dir = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'src', 'locales')
const load = (f) => JSON.parse(fs.readFileSync(path.join(dir, f), 'utf8'))

function flatten(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object') flatten(v, key, out)
    else out[key] = String(v)
  }
  return out
}

// Plural suffixes differ by language (en: _one/_other, others may add _few/_many…).
const PLURAL = /_(zero|one|two|few|many|other)$/
const base = (k) => k.replace(PLURAL, '')
const placeholders = (s) => [...s.matchAll(/\{\{\s*(\w+)\s*\}\}/g)].map((m) => m[1]).sort().join(',')

const ref = flatten(load('en.json'))
const refBases = new Set(Object.keys(ref).map(base))
// Keys whose English value is intentionally empty.
const allowEmpty = new Set(['errors.ERR_CANCELED.hint'])

let failed = false
for (const file of fs.readdirSync(dir).filter((f) => f.endsWith('.json') && f !== 'en.json')) {
  const loc = flatten(load(file))
  const problems = []
  const locBases = new Set(Object.keys(loc).map(base))
  for (const b of refBases) if (!locBases.has(b)) problems.push(`missing: ${b}`)
  for (const b of locBases) if (!refBases.has(b)) problems.push(`extra: ${b}`)
  for (const [k, v] of Object.entries(loc)) {
    if (!v.trim() && !allowEmpty.has(k)) problems.push(`empty: ${k}`)
    const refKey = k in ref ? k : Object.keys(ref).find((r) => base(r) === base(k) && r.endsWith('_other'))
    if (refKey && placeholders(ref[refKey]) !== placeholders(v)) {
      problems.push(`placeholders differ in ${k}: "${v}" vs en "${ref[refKey]}"`)
    }
  }
  if (problems.length) {
    failed = true
    console.error(`✗ ${file}\n  ${problems.join('\n  ')}`)
  } else {
    console.log(`✓ ${file} (${Object.keys(loc).length} strings)`)
  }
}
process.exit(failed ? 1 : 0)
