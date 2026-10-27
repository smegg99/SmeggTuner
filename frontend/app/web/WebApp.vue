<!-- app/web/WebApp.vue -->
<!-- The web edition's root (nuxt.config.ts swaps it in for app.vue): the engine loads behind the boot screen, then the
     window takes over. -->
<template>
  <v-app v-cloak>
    <!-- A cross-fade: the window fades in under the boot screen while it dissolves, so there is no empty frame between. -->
    <Transition name="boot">
      <WebAppWindow v-if="entered" />
      <ShellBootScreen
        v-else
        :failure="status === 'failed' ? failure || t('web.loadFailed') : undefined"
      >
        {{ t('web.loadFailed') }}
      </ShellBootScreen>
    </Transition>
  </v-app>
</template>

<script setup lang="ts">
import { useBootAppearance } from './bootAppearance'
import { useWebEngine } from './useWebEngine'

// The boot screen stays at least this long, so it reads as a deliberate start rather than a flicker.
const MIN_BOOT_MS = 3200

const { t, locale } = useI18n()
useHead(() => ({
  title: 'SmeggTuner',
  htmlAttrs: { lang: locale.value },
  link: [
    { rel: 'manifest', href: '/manifest.webmanifest' },
    { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/icons/icon-32.png' },
    { rel: 'apple-touch-icon', href: '/icons/icon-180.png' },
  ],
  meta: [
    { name: 'theme-color', content: '#000000' },
    { name: 'apple-mobile-web-app-capable', content: 'yes' },
    { name: 'mobile-web-app-capable', content: 'yes' },
    { name: 'apple-mobile-web-app-title', content: 'SmeggTuner' },
  ],
}))

useBootAppearance()

const { status, failure, boot } = useWebEngine()
const elapsed = ref(0)

// The window follows the engine on its own; the boot screen only holds for the minimum time.
const entered = computed(() => status.value === 'ready' && elapsed.value >= 1)

let frame = 0
onMounted(() => {
  const t0 = performance.now()
  const tick = () => {
    elapsed.value = Math.min(1, (performance.now() - t0) / MIN_BOOT_MS)
    if (elapsed.value < 1) frame = requestAnimationFrame(tick)
  }
  frame = requestAnimationFrame(tick)
  void boot()
  // Production only: a dev server's cache would outlive every edit.
  if (import.meta.env.PROD && 'serviceWorker' in navigator) void navigator.serviceWorker.register('/sw.js').catch(() => {})
})
onBeforeUnmount(() => cancelAnimationFrame(frame))
</script>

<style>
[v-cloak] {
  display: none;
}
</style>

<style scoped>
.boot-leave-active {
  transition: opacity 0.9s cubic-bezier(0.4, 0, 0.2, 1), transform 0.9s cubic-bezier(0.4, 0, 0.2, 1);
}
.boot-leave-to {
  opacity: 0;
  transform: scale(1.04);
}
.boot-enter-active { transition: opacity 0.7s ease-out 0.15s; }
.boot-enter-from { opacity: 0; }
@media (prefers-reduced-motion: reduce) {
  .boot-leave-active,
  .boot-enter-active { transition: opacity 0.3s linear; }
  .boot-leave-to { transform: none; }
}
</style>
