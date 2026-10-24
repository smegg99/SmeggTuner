// app/plugins/editKeys.ts
import { installDesktopEditKeys } from '~/composables/s99wails'

export default defineNuxtPlugin((app) => {
  const dispose = installDesktopEditKeys()
  app.vueApp.onUnmount(dispose)
  if (import.meta.hot) import.meta.hot.dispose(dispose)
})
