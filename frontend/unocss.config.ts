import { defineConfig, presetIcons } from 'unocss'
import { aliases } from 'vuetify/iconsets/mdi'

// Icons are UnoCSS masks named like the mdi-* classes, so only the icons in use ship. Names in
// TypeScript (utils, S99AB's toasts) are scanned too; Vuetify's own come from its alias table.
export default defineConfig({
  content: {
    pipeline: {
      include: [/\.(vue|[jt]sx?|html)($|\?)/],
    },
  },
  safelist: Object.values(aliases).filter((name): name is string => typeof name === 'string'),
  presets: [
    presetIcons({
      prefix: '',
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle',
      },
    }),
  ],
})
