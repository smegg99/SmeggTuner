// i18n/i18n.config.ts
import { polishPlural } from './plurals'

export default defineI18nConfig(() => ({
  legacy: false,
  fallbackLocale: 'en',

  missingWarn: import.meta.dev,
  fallbackWarn: import.meta.dev,

  // English keeps vue-i18n's own rule; only Polish needs telling.
  pluralRules: { pl: polishPlural },

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
