// The Smegglect overlay for reporting UI issues. nuxt.config.ts registers this plugin only for builds made with
// SMEGGLECT_DEBUG=1 (debug builds in scripts/build.sh); a production build never sees it. Sent reports reach
// Claude Code through the Smegglect service's Bridge and the broker `just dev` starts.

import { mount } from 'smegglect'
import { routeContext } from 'smegglect/vue'
import { bindingBridge, bindingSave } from 'smegglect/wails'
import { useShell } from '~/composables/useShell'
import * as SmegglectService from '~~bindings/smegg.me/smeggtuner/backend/services/smegglect/service.js'

export default defineNuxtPlugin((nuxt) => {
  const router = useRouter()
  const { view } = useShell()
  mount({
    project: { id: 'smeggtuner', name: 'SmeggTuner' },
    // The overlay speaks the window's language, not the WebView's.
    locale: nuxt.vueApp.config.globalProperties.$i18n.locale,
    context: [routeContext(router)],
    // One route: the views are the shell's state, so an issue or region belongs to the view it was made in.
    screen: () => view.value,
    save: bindingSave((name, data) => SmegglectService.Save(name, data)),
    bridge: { label: 'SmeggTuner', via: bindingBridge(request => SmegglectService.Bridge(request)) },
  })
})
