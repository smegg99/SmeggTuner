// app/composables/useThemeSync.ts
import { computed, onMounted, onScopeDispose, watch } from 'vue'
import { useLocale, useTheme } from 'vuetify'
import { registerAppearanceSource, resetAppearanceSource, useAppearance } from '@smegg99/s99ab/appearance'
import { themeState } from '@smegg99/s99ab/theme'
import * as ThemeService from '~~bindings/github.com/smegg99/s99wails/v2/services/theme/service.js'
import { resolveDesktopTheme, useSystemTheme, type SystemThemeInfo } from '~/composables/s99wails'
import { applyInstrumentTheme, TOKENS } from '~/theme/tokens'
import { accentMode, themeMode } from '~/types/config'
import { WEB, WEB_ACCENT } from '~/utils/features'
import type { LocaleCode } from '~/types/locale'
import type { AccentMode, ThemeMode } from '~/types/config'

// Called once by the root so the native listener and shared appearance have one owner.
export function initializeThemeSync() {
  const { config } = useConfigSync()
  const { locale, setLocale } = useI18n()
  const vuetifyLocale = useLocale()
  watch(locale, (code) => {
    vuetifyLocale.current.value = code
  }, { immediate: true })
  const log = useLogger()
  const { theme } = useSystemTheme({
    getTheme: async (): Promise<SystemThemeInfo> => {
      try {
        const info = await ThemeService.GetTheme()
        return { ...info, scheme: info.scheme || 'no-preference' }
      }
      catch (error) {
        log.error('log.frontendError', { error: String(error) })
        return { accent_color: '', dark_mode: false, scheme: 'no-preference', accent_watch_supported: false }
      }
    },
  })

  registerAppearanceSource({
    dark: () => theme.value?.dark_mode,
  })
  onScopeDispose(resetAppearanceSource)
  const appearance = useAppearance(() => ({ theme: 'smeggtuner.theme' }))
  const vuetify = useTheme()
  const desktop = resolveDesktopTheme({
    system: theme,
    mode: () => themeMode(config.preferences.theme),
    accentMode: () => accentMode(config.preferences.accent_mode),
    customAccent: () => config.preferences.accent_color,
    fallbackAccent: dark => WEB ? WEB_ACCENT : TOKENS[dark ? 'dark' : 'light'].accent,
  })
  // S99Wails resolves the native preferences; this instrument keeps its established palette.
  watch(() => [desktop.accentFor(false), desktop.accentFor(true), themeState().high], () => {
    for (const name of ['light', 'dark'] as const) {
      applyInstrumentTheme(vuetify.themes.value[name]!, name, desktop.accentFor(name === 'dark'))
    }
  }, { immediate: true, flush: 'post' })

  onMounted(() => {
    watch(() => config.preferences.theme, (value) => {
      const mode = themeMode(value)
      appearance.setMode(mode === 'auto' ? 'system' : mode, undefined, 'none')
    }, { immediate: true })
    watch(() => config.preferences.language, async (language) => {
      const code = language === 'pl' ? 'pl' : 'en'
      if (locale.value !== code) await setLocale(code)
    }, { immediate: true })
  })
}

export function useThemeSync() {
  const { config, saveConfigNow } = useConfigSync()
  const theme = useTheme()
  const { locale, locales, setLocale } = useI18n()
  const localeItems = computed(() =>
    locales.value.map(item => ({ code: item.code as LocaleCode, name: item.name ?? item.code })),
  )

  async function setLanguage(code: LocaleCode) {
    config.preferences.language = code
    if (locale.value !== code) await setLocale(code)
    await saveConfigNow()
  }

  return {
    prefs: computed(() => config.preferences),
    mode: computed(() => themeMode(config.preferences.theme)),
    isDark: computed(() => theme.global.current.value.dark),
    accent: computed(() => accentMode(config.preferences.accent_mode)),
    accentColor: computed(() => config.preferences.accent_color),
    localeItems,
    setLanguage,
    setThemeMode: (mode: ThemeMode) => { config.preferences.theme = mode },
    setAccentMode: (mode: AccentMode) => { config.preferences.accent_mode = mode },
    setAccentColor: (color: string) => { config.preferences.accent_color = color },
  }
}
