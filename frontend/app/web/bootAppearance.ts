// app/web/bootAppearance.ts
// The boot screen shows before the engine, so before initializeThemeSync; it applies the same theme and language from
// the stored settings so nothing flips when the window takes over.
import { useTheme } from 'vuetify'
import { useAppearance } from '@smegg99/s99ab/appearance'
import { applyInstrumentTheme } from '~/theme/tokens'
import { WEB_ACCENT } from '~/utils/features'
import { loadConfig, storedPreferences } from './storage'

export function useBootAppearance() {
  const { theme, language } = storedPreferences(loadConfig())
  const vuetify = useTheme()
  for (const name of ['light', 'dark'] as const) {
    applyInstrumentTheme(vuetify.themes.value[name]!, name, WEB_ACCENT)
  }
  const dark = theme === 'dark' || (theme !== 'light' && typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches)
  useAppearance(() => ({ theme: 'smeggtuner.theme' })).setMode(dark ? 'dark' : 'light', undefined, 'none')
  const { locale, setLocale } = useI18n()
  const code = language === 'pl' ? 'pl' : language === 'en' ? 'en' : ''
  if (code && locale.value !== code) void setLocale(code)
}
