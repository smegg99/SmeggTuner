import type { ThemeDefinition } from 'vuetify'
import { TOKENS } from '~/theme/tokens'

// Only the Vuetify components this app renders are configured. No preset carries a `color`:
// the accent is a token read through CSS, and a component default is how a stray blue would get in.
export const DEFAULTS = {
  VBtn: { variant: 'flat', elevation: 0, density: 'compact', rounded: 'sm' },
  VCard: { flat: true, elevation: 0, rounded: 'sm', border: true },
  VDialog: { scrollable: true },
  VTooltip: { location: 'bottom' },
  VMenu: { location: 'bottom' },
} as const

/**
 * themes maps our tokens onto Vuetify's names and registers each token under its own name too.
 *
 * Footgun: Vuetify emits `--v-theme-<name>` as a comma-separated triplet (`21,23,27`), not a colour.
 * Use `rgb(var(--v-theme-ink))` or `rgba(var(--v-theme-neutral), 0.22)`; the slash-alpha form
 * `rgb(var(--x) / 0.22)` compiles to invalid CSS and is dropped in silence (no background, no console error).
 */
export function themes(): Record<'light' | 'dark', ThemeDefinition> {
  const build = (name: 'light' | 'dark'): ThemeDefinition => {
    const t = TOKENS[name]

    return {
      dark: name === 'dark',
      colors: {
        ...t,

        // Vuetify's own names, so its components land in our palette.
        'background': t.bg,
        'surface': t.chrome,
        'surface-bright': t.raised,
        'surface-light': t.chrome2,
        'surface-variant': t.sunk,
        'on-background': t.ink,
        'on-surface': t.ink,

        // Vuetify requires these; they are mapped onto our tokens.
        'primary': t.ink2,
        'secondary': t.ink3,
        'error': t.reed,
        'success': t.goal,
        'warning': t.warn,
        'info': t.ink2,
      },
      variables: {
        'border-color': t.line,
        'border-opacity': 1,
        'high-emphasis-opacity': 1,
        'medium-emphasis-opacity': 1,
      },
    }
  }

  return { light: build('light'), dark: build('dark') }
}
