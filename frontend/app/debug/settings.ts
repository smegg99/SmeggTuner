// The settings a debug build adds to the Settings room. nuxt.config.ts registers this plugin beside the Smegglect
// overlay, for builds made with SMEGGLECT_DEBUG=1 only; a production build never sees it.

import { debugSettings } from '~/utils/debugSettings'
import PhoneLayout from './PhoneLayout.vue'
import { restorePhoneLayout } from './phone'

export default defineNuxtPlugin(() => {
  debugSettings.value = PhoneLayout
  void restorePhoneLayout()
})
