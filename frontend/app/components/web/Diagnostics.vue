<!-- app/components/web/Diagnostics.vue -->
<template>
  <UiFormSection
    :title="t('web.diagnostics.title')"
    class="tuner-section settings__section"
  >
    <p
      v-if="!track"
      class="diag__empty"
    >
      {{ t('web.diagnostics.empty') }}
    </p>
    <pre class="diag">{{ text }}</pre>
    <v-btn
      variant="tonal"
      prepend-icon="mdi-content-copy"
      @click="copy"
    >
      {{ copied ? t('web.diagnostics.copied') : t('web.diagnostics.copy') }}
    </v-btn>
  </UiFormSection>
</template>

<script setup lang="ts">
import { diagnosticsText } from '~/web/diagnostics'
import { currentPlatform } from '~/web/platform'
import { useWebEngine } from '~/web/useWebEngine'

const { t } = useI18n()
const { track, perf, context } = useWebEngine()
const { inputLevel } = useTuner()
const version = useRuntimeConfig().public.version
const copied = ref(false)

const text = computed(() => {
  const ctx = context()
  return diagnosticsText({
    version,
    userAgent: navigator.userAgent,
    platform: currentPlatform(),
    track: track.value,
    contextRate: ctx?.sampleRate ?? null,
    baseLatency: ctx?.baseLatency ?? null,
    inputLevel: inputLevel.value ?? null,
    perf: perf.value,
  })
})

async function copy() {
  try {
    await navigator.clipboard.writeText(text.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  }
  catch {
    copied.value = false
  }
}
</script>

<style scoped>
.diag { font-family: var(--font-mono); font-size: 11px; margin: 0 0 8px; overflow-x: auto; white-space: pre-wrap; overflow-wrap: anywhere; }
.diag__empty { color: rgb(var(--v-theme-ink3)); font-size: 12px; margin: 0 0 6px; }
</style>
