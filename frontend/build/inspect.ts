// build/inspect.ts
import type { AddressInfo } from 'node:net'
import { addPluginTemplate, defineNuxtModule } from 'nuxt/kit'
import { startBroker } from 'smegglect/bridge'
import type { Broker } from 'smegglect/bridge'

// Smegglect for `nuxt dev` only: the overlay, and its agent bridge for the address the dev server
// really listens at, which is known only once it does. No build loads this module.
export default defineNuxtModule({
  meta: { name: 'smeggtuner-inspect' },
  setup(_options, nuxt) {
    let broker: Broker | undefined
    nuxt.hook('listen', async (server) => {
      const { port } = server.address() as AddressInfo
      broker = await startBroker({
        root: nuxt.options.rootDir,
        origins: ['127.0.0.1', 'localhost'].map(host => `http://${host}:${port}`),
      })
    })
    nuxt.hook('close', () => broker?.close())

    addPluginTemplate({
      filename: 'smegglect.client.mjs',
      mode: 'client',
      // Written after the server listens, so the bridge's address and token are in it.
      getContents: () => `import { defineNuxtPlugin, useRouter } from '#app'
import { mount } from 'smegglect'
import { routeContext, routeScreen, vueSource } from 'smegglect/vue'

export default defineNuxtPlugin(() => {
  const router = useRouter()
  const overlay = mount({
    project: { id: 'smeggtuner', name: 'SmeggTuner' },
    bridge: ${JSON.stringify(broker && { url: broker.url, token: broker.token })},
    sources: [vueSource({ root: ${JSON.stringify(nuxt.options.rootDir)} })],
    context: [routeContext(router)],
    screen: routeScreen(router),
  })
  if (import.meta.hot) import.meta.hot.dispose(() => overlay.destroy())
})
`,
    })
  },
})
