// app/plugins/00.browser.ts
import 'core-js/actual/array/to-sorted'
import 'core-js/actual/array/to-reversed'
import 'core-js/actual/array/to-spliced'
import 'core-js/actual/array/find-last'
import 'core-js/actual/array/find-last-index'
import { flattenLayers, THEME_STYLESHEET_ID } from './browser/layers'

export default defineNuxtPlugin({
  name: 'browser',
  enforce: 'pre',
  dependsOn: ['nuxt:head'],
  setup(app) {
    const head = injectHead()
    if (!head?.hooks) throw new Error('Nuxt head hooks are unavailable')
    // Vuetify exposes its theme through Unhead but has no option to disable CSS layers.
    const dispose = head.hooks.hook('tags:resolve', ({ tags }) => {
      for (const tag of tags) {
        if (tag.tag !== 'style' || tag.props.id !== THEME_STYLESHEET_ID) continue
        if (typeof tag.textContent === 'string') tag.textContent = flattenLayers(tag.textContent)
        if (typeof tag.innerHTML === 'string') tag.innerHTML = flattenLayers(tag.innerHTML)
      }
    })
    app.vueApp.onUnmount(dispose)
  },
})
