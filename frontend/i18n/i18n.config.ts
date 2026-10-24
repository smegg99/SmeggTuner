// i18n/i18n.config.ts
import { withS99ab } from '@smegg99/s99ab/i18n'
import { pluralRules } from './generated/plurals.gen'

export default defineI18nConfig(() => ({
  legacy: false,
  messages: withS99ab({ en: {}, pl: {} }),
  fallbackLocale: 'en',

  missingWarn: import.meta.dev,
  fallbackWarn: import.meta.dev,

  pluralRules,

  datetimeFormats: {
    en: {
      short: { year: 'numeric', month: 'short', day: '2-digit' },
      medium: { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' },
    },
    pl: {
      short: { year: 'numeric', month: 'short', day: '2-digit' },
      medium: { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' },
    },
  },
}))
