import { readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const versionFile = path.join(root, 'VERSION')
const args = process.argv.slice(2)
const check = args[0] === '--check'
const requested = check ? undefined : args[0]

if (args.length > 1) {
  throw new Error('usage: node scripts/sync-version.mjs [--check|VERSION]')
}

const version = requested
  ? requested.replace(/^v/, '')
  : readFileSync(versionFile, 'utf8').trim()

if (!/^\d+\.\d+\.\d+$/.test(version)) {
  throw new Error(`invalid version "${version}": expected MAJOR.MINOR.PATCH`)
}

if (requested) {
  writeFileSync(versionFile, `${version}\n`)
}

const files = [
  {
    path: 'build/config.yml',
    replacements: [/(version: ")[0-9]+\.[0-9]+\.[0-9]+(" # The application version)/],
  },
  {
    path: 'build/darwin/Info.dev.plist',
    replacements: [
      /(<key>CFBundleShortVersionString<\/key>\s*<string>)[^<]+(<\/string>)/,
      /(<key>CFBundleVersion<\/key>\s*<string>)[^<]+(<\/string>)/,
    ],
  },
  {
    path: 'build/darwin/Info.plist',
    replacements: [
      /(<key>CFBundleShortVersionString<\/key>\s*<string>)[^<]+(<\/string>)/,
      /(<key>CFBundleVersion<\/key>\s*<string>)[^<]+(<\/string>)/,
    ],
  },
  {
    path: 'build/windows/info.json',
    replacements: [
      /("file_version": ")[0-9]+\.[0-9]+\.[0-9]+(")/,
      /("ProductVersion": ")[0-9]+\.[0-9]+\.[0-9]+(")/,
    ],
  },
  {
    path: 'build/windows/nsis/wails_tools.nsh',
    replacements: [/(!define INFO_PRODUCTVERSION ")[0-9]+\.[0-9]+\.[0-9]+(")/],
  },
  {
    path: 'build/windows/wails.exe.manifest',
    replacements: [/(name="me\.smegg\.smeggtuner" version=")[0-9]+\.[0-9]+\.[0-9]+(")/],
  },
  {
    path: 'build/windows/msix/app_manifest.xml',
    replacements: [/(<Identity\s[\s\S]*?\bVersion=")[0-9]+\.[0-9]+\.[0-9]+\.0(")/],
    suffix: '.0',
  },
  {
    path: 'build/windows/msix/template.xml',
    replacements: [/(<PackageInformation\s[\s\S]*?\bVersion=")[0-9]+\.[0-9]+\.[0-9]+\.0(")/],
    suffix: '.0',
  },
]

const stale = []

for (const file of files) {
  const filePath = path.join(root, file.path)
  const source = readFileSync(filePath, 'utf8')
  let updated = source

  for (const pattern of file.replacements) {
    const matches = updated.match(new RegExp(pattern.source, 'g')) ?? []
    if (matches.length !== 1) {
      throw new Error(`${file.path}: expected one version field matching ${pattern}, found ${matches.length}`)
    }
    updated = updated.replace(pattern, (_, before, after) => `${before}${version}${file.suffix ?? ''}${after}`)
  }

  if (updated === source) {
    continue
  }
  if (check) {
    stale.push(file.path)
  }
  else {
    writeFileSync(filePath, updated)
  }
}

if (stale.length > 0) {
  throw new Error(`version metadata is stale:\n${stale.map(file => `  ${file}`).join('\n')}`)
}

console.log(`${check ? 'checked' : 'synced'} version ${version}`)
