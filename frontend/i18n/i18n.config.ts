// i18n/i18n.config.ts
import { withS99ab } from '@smegg99/s99ab/i18n'
import { locales } from './generated/keys.gen'
import { pluralRules } from './generated/plurals.gen'

function perLocale<T>(value: T) {
  return Object.fromEntries(locales.map(code => [code, value])) as Record<typeof locales[number], T>
}

export default defineI18nConfig(() => ({
  legacy: false,
  messages: withS99ab(perLocale({})),
  fallbackLocale: 'en',
  missingWarn: import.meta.dev,
  fallbackWarn: import.meta.dev,
  pluralRules,
  datetimeFormats: perLocale({
    short: { year: 'numeric', month: 'short', day: '2-digit' },
    medium: { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' },
  } as const),
}))
