import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import { DEFAULT_THEME_NAME } from '~/types/config'
import { DEFAULTS, themes } from '~/theme/defaults'

export default defineNuxtPlugin((app) => {
  app.vueApp.use(createVuetify({
    theme: {
      defaultTheme: DEFAULT_THEME_NAME,
      themes: themes(),
    },
    defaults: DEFAULTS as never,
  }))
})
