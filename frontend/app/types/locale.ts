// app/types/locale.ts
import { defaultLocale, locales } from '../../i18n/generated/keys.gen'

export type LocaleCode = typeof locales[number]

export interface LocaleInfo {
  code: LocaleCode
  name: string
  language: string
  files: string[]
}

export const LOCALE_CODES = locales

export const DEFAULT_LOCALE: LocaleCode = defaultLocale
