// app/plugins/01.viewport.client.ts
import { desktopScale } from '~/utils/desktopViewport'

export default defineNuxtPlugin({
  name: 'desktop-viewport',
  dependsOn: ['browser'],
  setup(app) {
    const content = ref('width=device-width, initial-scale=1')
    useHead(() => ({ meta: [{ name: 'viewport', content: content.value }] }))
    const viewport = window.visualViewport
    let frame = 0

    function fit() {
      // Multiplication recovers the available screen size after the previous fit.
      const width = Math.round(viewport ? viewport.width * viewport.scale : window.innerWidth)
      const height = Math.round(viewport ? viewport.height * viewport.scale : window.innerHeight)
      const scale = desktopScale(width, height)
      // Keep Android's automatic text-field zoom from cropping the fitted interface.
      content.value = `width=960, initial-scale=${scale}, maximum-scale=${scale}`
    }
    function resize() {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(fit)
    }

    fit()
    window.addEventListener('resize', resize)
    viewport?.addEventListener('resize', resize)
    app.vueApp.onUnmount(() => {
      cancelAnimationFrame(frame)
      window.removeEventListener('resize', resize)
      viewport?.removeEventListener('resize', resize)
    })
  },
})
