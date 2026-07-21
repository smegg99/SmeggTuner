import { computed, effectScope, onMounted, shallowRef, watch } from 'vue'
import type { Ref } from 'vue'
import * as ThemeService from '~~bindings/github.com/smegg99/s99wails/services/theme/service.js'
import { Info as ThemeInfo } from '~~bindings/github.com/smegg99/s99wails/accent/models.js'
import { useTheme } from 'vuetify'
import { useSystemTheme, type SystemThemeInfo } from '~/composables/s99wails'
import { accentMode, resolveThemeName, themeMode } from '~/types/config'
import { TOKENS } from '~/theme/tokens'
import { resolveDesktopTheme, type DesktopTheme } from '~/composables/s99wails/resolveDesktopTheme'
import type { LocaleCode } from '~/types/locale'
import type { AccentMode, ThemeMode } from '~/types/config'

// useSystemTheme registers a listener per call and five components use this, so it is consumed once in start() and mirrored into a module ref (also giving dark_mode a value before start() runs).

const systemTheme = shallowRef<SystemThemeInfo | null>(null)
const darkMode = computed(() => systemTheme.value?.dark_mode ?? false)
const scope = effectScope(true)

let started = false

type LocaleSetter = (code: LocaleCode) => Promise<void> | void

interface ThemeConfig {
  preferences: {
    theme: string
    language: string
    accent_mode: string
    accent_color: string
  }
}

// Route the accent through Vuetify's theme (not a raw --v-theme-accent) so genOnColors regenerates readable on-accent text.
function applyVuetifyAccent(vuetifyTheme: ReturnType<typeof useTheme>, desktop: DesktopTheme) {
  for (const name of ['light', 'dark'] as const) {
    const theme = vuetifyTheme.themes.value[name]
    if (theme) theme.colors.accent = desktop.accentFor(name === 'dark')
  }
}

async function applyLocale(locale: Ref<string>, setLocale: LocaleSetter, code: LocaleCode) {
  if (locale.value !== code) {
    await setLocale(code)
  }
}

function start(options: {
  config: ThemeConfig
  vuetifyTheme: ReturnType<typeof useTheme>
  locale: Ref<string>
  setLocale: LocaleSetter
  log: ReturnType<typeof useLogger>
}) {
  if (started) return
  started = true

  // Module-scoped so the listener lives as long as the app, not the first component to mount.
  const { theme } = scope.run(() =>
    useSystemTheme({
      // Not redundant: useSystemTheme does not handle a rejection, and this fails every time in browser dev (no Go side), so log and fall back to defaults.
      getTheme: () => ThemeService.GetTheme().catch((err) => {
        options.log.error('failed to load theme info', { error: String(err) })
        return new ThemeInfo() as SystemThemeInfo
      }),
    }),
  )!

  scope.run(() => {
    watch(theme, (info) => {
      systemTheme.value = info
    }, { immediate: true })

    watch(() => options.config.preferences.language, (language) => {
      void applyLocale(options.locale, options.setLocale, (language || 'en') as LocaleCode)
    }, { immediate: true })

    // The resolver's getters read preferences fresh: useConfigSync replaces the whole preferences object each sync, so a captured reference goes dead.
    const desktop = resolveDesktopTheme({
      system: systemTheme,
      mode: () => themeMode(options.config.preferences.theme),
      accentMode: () => accentMode(options.config.preferences.accent_mode),
      customAccent: () => options.config.preferences.accent_color,
      fallbackAccent: dark => TOKENS[dark ? 'dark' : 'light'].accent,
    })
    watch(desktop.themeName, (name) => {
      options.vuetifyTheme.global.name.value = name
    }, { immediate: true })
    watch(desktop.accent, () => applyVuetifyAccent(options.vuetifyTheme, desktop), { immediate: true })
  })
}

export function useThemeSync() {
  const { config, saveConfigNow } = useConfigSync()
  const vuetifyTheme = useTheme()
  const { locale, locales, setLocale } = useI18n()
  const log = useLogger()

  const localeItems = computed(() =>
    (locales.value as Array<{ code: string, name: string }>).map(l => ({
      code: l.code as LocaleCode,
      name: l.name,
    })),
  )

  const prefs = computed(() => config.preferences)
  const mode = computed<ThemeMode>(() => themeMode(config.preferences?.theme))
  const isDark = computed(() => resolveThemeName(mode.value, darkMode.value) === 'dark')
  const accent = computed<AccentMode>(() => accentMode(config.preferences?.accent_mode))
  const accentColor = computed(() => config.preferences?.accent_color ?? '')

  onMounted(() => {
    start({ config, vuetifyTheme, locale, setLocale, log })
  })

  function setThemeMode(next: ThemeMode) {
    config.preferences.theme = next
  }

  async function setLanguage(code: LocaleCode) {
    config.preferences.language = code
    await applyLocale(locale, setLocale, code)
    await saveConfigNow()
  }

  function setAccentMode(next: AccentMode) {
    config.preferences.accent_mode = next
  }

  function setAccentColor(hex: string) {
    config.preferences.accent_color = hex
  }

  return {
    prefs, mode, isDark, localeItems, setThemeMode, setLanguage,
    accent, accentColor, setAccentMode, setAccentColor,
  }
}
