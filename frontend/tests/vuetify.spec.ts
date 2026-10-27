import { describe, expect, it } from 'vitest'
import { createVuetify } from 'vuetify'
import { rebuildThemes, s99abTheme } from '@smegg99/s99ab/theme'
import { applyInstrumentTheme, DARK, LIGHT } from '~/theme/tokens'

describe('instrument appearance on the shared theme', () => {
  it('keeps the established surfaces and measurement colours after a shared repaint', () => {
    const vuetify = createVuetify(s99abTheme())
    rebuildThemes(vuetify.theme.themes.value, { accent: '#996633', high: true })
    for (const [name, tokens] of Object.entries({ light: LIGHT, dark: DARK })) {
      applyInstrumentTheme(vuetify.theme.themes.value[name]!, name as 'light' | 'dark', '#ff5c78')
      const colors = vuetify.theme.computedThemes.value[name]!.colors
      for (const [key, value] of Object.entries(tokens)) {
        expect(colors[key]).toBe(key === 'accent' ? '#ff5c78' : value)
      }
      expect(colors.background).toBe(tokens.bg)
      expect(colors.surface).toBe(tokens.chrome)
      expect(colors.error).toBe(tokens.reed)
      expect(colors.success).toBe(tokens.goal)
    }
  })

  it('derives legible foregrounds again when the custom accent changes', () => {
    const vuetify = createVuetify(s99abTheme())
    for (const name of ['light', 'dark'] as const) {
      applyInstrumentTheme(vuetify.theme.themes.value[name]!, name, '#000000')
      expect(vuetify.theme.computedThemes.value[name]!.colors['on-accent']).toBe('#FFF')
      applyInstrumentTheme(vuetify.theme.themes.value[name]!, name, '#ffffff')
      expect(vuetify.theme.computedThemes.value[name]!.colors['on-accent']).toBe('#000')
      expect(vuetify.theme.computedThemes.value[name]!.colors.well).toBe((name === 'dark' ? DARK : LIGHT).well)
    }
  })
})
