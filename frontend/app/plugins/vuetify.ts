// app/plugins/vuetify.ts
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import {
  cs, da, de, el, en, es, et, fi, fr, hu, it, ja, ko, lt, lv, nl, no as nb, pl, pt, ro, ru, sk, sv, th, zhHans, zhHant,
} from 'vuetify/locale'
import { s99abTheme } from '@smegg99/s99ab/theme'
import { CheckTick } from '~/theme/checkTick'
import { applyInstrumentTheme } from '~/theme/tokens'

export default defineNuxtPlugin({
  name: 'vuetify',
  dependsOn: ['browser'],
  setup(app) {
    const options = s99abTheme()
    options.locale = {
      locale: 'en',
      fallback: 'en',
      messages: {
        cs, da, de, el, en, es, et, fi, fr, hu, it, ja, ko, lt, lv, nb, nl, pl, ro, ru, sk, sv, th,
        'pt-BR': pt, 'pt-PT': pt, 'zh-Hans': zhHans, 'zh-Hant': zhHant,
      },
    }
    options.defaults = {
      ...options.defaults,
      VBtn: { variant: 'flat', elevation: 0, density: 'compact' },
      VCheckbox: { color: 'ink2', density: 'compact', hideDetails: 'auto', trueIcon: CheckTick, falseIcon: '' },
      VSlider: { color: 'ink2', density: 'compact', hideDetails: true, thumbSize: 10, trackSize: 3 },
      VChip: { variant: 'outlined', size: 'small', rounded: false },
    }
    const vuetify = createVuetify(options)
    for (const name of ['light', 'dark'] as const) {
      applyInstrumentTheme(vuetify.theme.themes.value[name]!, name)
    }
    app.vueApp.use(vuetify)
  },
})
