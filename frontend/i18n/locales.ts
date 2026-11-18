// i18n/locales.ts
import { defaultLocale, locales } from './generated/keys.gen'

export function isLocale(value: unknown): value is typeof locales[number] {
  return typeof value === 'string' && (locales as readonly string[]).includes(value)
}

export function savedLocale(value: unknown): typeof locales[number] {
  return isLocale(value) ? value : defaultLocale
}

const regionalTags: Record<string, string> = { en: 'en-GB', pl: 'pl-PL' }

export const localeOptions = locales.map(code => ({
  code,
  language: regionalTags[code] ?? code,
  name: code === 'pl' ? 'Polski' : new Intl.DisplayNames([code], { type: 'language' }).of(code) ?? code,
}))
