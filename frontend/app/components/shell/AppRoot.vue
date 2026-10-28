<!-- app/components/shell/AppRoot.vue -->
<!-- The desktop app: the window, with the boot screen dissolving off it. -->
<template>
  <v-app v-cloak>
    <ShellAppWindow />
    <!-- The window starts underneath; the boot screen dissolves off it. -->
    <Transition name="boot">
      <ShellBootScreen v-if="booting" />
    </Transition>
  </v-app>
</template>

<script setup lang="ts">
import * as LoggerService from '~~bindings/github.com/smegg99/s99wails/v2/services/logger/service.js'
import { installCrashHeartbeat, installFrontendLogBridge } from '~/composables/s99wails'
import { useConfigSync } from '~/composables/useConfigSync'
import { initializeThemeSync } from '~/composables/useThemeSync'
import { filterResizeNotice } from '~/utils/resizeNotice'

// How long the boot screen holds before it fades: a short, deliberate start, not the web's engine wait.
const BOOT_MS = 700

// The window is mounted above the router so a route change never tears down a running engine's subscriptions; rooms are switched by state, not routes.
useConfigSync()
initializeThemeSync()

// Routes uncaught errors and unhandled rejections to the Go log; installed for the window's life, a no-op outside the desktop.
const { locale } = useI18n()
useHead(() => ({ htmlAttrs: { lang: locale.value } }))
const log = useLogger()
const disposeNotice = filterResizeNotice(() => log.debug('log.frontendResizeNotice'))
// Tells the crash watcher the page still runs; a page that stops opens the not-responding screen.
const stopHeartbeat = installCrashHeartbeat()
const disposeLog = installFrontendLogBridge(LoggerService, { error: 'log.frontendError', rejection: 'log.frontendRejection' })
const booting = ref(true)
const bootTimer = setTimeout(() => {
  booting.value = false
}, BOOT_MS)
onBeforeUnmount(() => {
  clearTimeout(bootTimer)
  stopHeartbeat()
  disposeLog()
  disposeNotice()
})
</script>

<style>
[v-cloak] {
  display: none;
}

.boot-leave-active {
  transition: opacity 0.35s cubic-bezier(0.4, 0, 0.2, 1), transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}
.boot-leave-to {
  opacity: 0;
  transform: scale(1.04);
}
@media (prefers-reduced-motion: reduce) {
  .boot-leave-active { transition: opacity 0.2s linear; }
  .boot-leave-to { transform: none; }
}
</style>
