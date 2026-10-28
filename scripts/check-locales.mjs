// scripts/check-locales.mjs
import { execFileSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const root = fileURLToPath(new URL('../', import.meta.url))
const scan = JSON.parse(execFileSync('go', ['tool', 'loc', 'lint', 'keys', '--json', '-c', 'i18n/config.cue'], {
  cwd: root,
  encoding: 'utf8',
}))
if (!scan.filesScanned) throw new Error('Locale scanner did not inspect any source files')
if (scan.missing?.length) throw new Error(`Missing locale keys: ${JSON.stringify(scan.missing)}`)

// S99Locale's source scanner does not read Go HTML templates.
const template = readFileSync(new URL('../backend/core/report/templates/report.gohtml', import.meta.url), 'utf8')
for (const locale of ['en', 'pl']) {
  const messages = JSON.parse(readFileSync(new URL(`../backend/common/i18n/locales/${locale}.json`, import.meta.url), 'utf8'))
  for (const [, key] of template.matchAll(/\bT(?:f|n)?\s+"([\w.]+)"/g)) {
    if (!(key in messages)) throw new Error(`Missing ${locale} report key: ${key}`)
  }
}
console.log(`Locale references checked in ${scan.filesScanned} source files and the report template`)
