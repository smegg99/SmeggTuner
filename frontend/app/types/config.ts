/** The theme the user asked for. Older config files may hold high-contrast names; those normalize back to their base (see themeMode). */
export const THEME_MODE = {
  AUTO: 'auto',
  LIGHT: 'light',
  DARK: 'dark',
} as const

export type ThemeMode = typeof THEME_MODE[keyof typeof THEME_MODE]

/** The two themes Vuetify itself ships, and the two this app renders. */
export type ThemeName = 'light' | 'dark'

export const DEFAULT_THEME_MODE: ThemeMode = THEME_MODE.AUTO
export const DEFAULT_THEME_NAME: ThemeName = 'light'

export const THEME_MODES: { value: ThemeMode, icon: string, labelKey: string }[] = [
  { value: THEME_MODE.AUTO, icon: 'mdi-monitor', labelKey: 'settings.themes.auto' },
  { value: THEME_MODE.LIGHT, icon: 'mdi-white-balance-sunny', labelKey: 'settings.themes.light' },
  { value: THEME_MODE.DARK, icon: 'mdi-moon-waning-crescent', labelKey: 'settings.themes.dark' },
]

/** themeMode reads the config's theme field, tolerating whatever it holds. */
export function themeMode(value: string | undefined): ThemeMode {
  switch (value) {
    case THEME_MODE.LIGHT:
    case 'lightHighContrast':
      return THEME_MODE.LIGHT
    case THEME_MODE.DARK:
    case 'darkHighContrast':
      return THEME_MODE.DARK
    case THEME_MODE.AUTO:
      return THEME_MODE.AUTO
    default:
      return DEFAULT_THEME_MODE
  }
}

/** resolveThemeName is the Vuetify theme to show, given the mode and the OS. */
export function resolveThemeName(mode: ThemeMode, systemDark: boolean): ThemeName {
  if (mode === THEME_MODE.AUTO) return systemDark ? 'dark' : 'light'
  return mode
}

/** How the accent is chosen. common/config #AccentMode. AUTO follows the desktop accent (s99wails reads it off the OS), falling back to the app's blue; CUSTOM uses the picked colour. */
export const ACCENT_MODE = {
  AUTO: 'auto',
  CUSTOM: 'custom',
} as const

export type AccentMode = typeof ACCENT_MODE[keyof typeof ACCENT_MODE]

export const DEFAULT_ACCENT_MODE: AccentMode = ACCENT_MODE.AUTO

/** The blue the app ships with, and what CUSTOM starts from. Mirrors tokens.ts. */
export const DEFAULT_ACCENT_COLOR = '#2563eb'

export const ACCENT_MODES: { value: AccentMode, icon: string, labelKey: string }[] = [
  { value: ACCENT_MODE.AUTO, icon: 'mdi-monitor', labelKey: 'settings.accent.auto' },
  { value: ACCENT_MODE.CUSTOM, icon: 'mdi-palette-outline', labelKey: 'settings.accent.custom' },
]

/** accentMode reads the config's accent_mode field, tolerating whatever it holds. */
export function accentMode(value: string | undefined): AccentMode {
  return value === ACCENT_MODE.CUSTOM ? ACCENT_MODE.CUSTOM : ACCENT_MODE.AUTO
}

/** Which error convention the tuner shows. core/target.Reference / tuner.error_reference. SCALE shows distance from the tempered scale, GOAL distance from the curve; the backend computes both, neither is derived here. */
export const ERROR_REFERENCE = {
  SCALE: 'scale',
  GOAL: 'goal',
} as const

export type ErrorReference = typeof ERROR_REFERENCE[keyof typeof ERROR_REFERENCE]

export const DEFAULT_ERROR_REFERENCE: ErrorReference = ERROR_REFERENCE.SCALE

/** Reed/beat windows in cents (common/config schema defaults = core/target Default*Tolerance). Used only before config loads; every inTol verdict is the backend's, these just size the gauge's in-tune band. */
export const DEFAULT_TOLERANCE = 1.0
export const DEFAULT_BEAT_TOLERANCE = 3.0
