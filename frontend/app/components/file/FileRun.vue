<template>
  <span class="run">
    <UiToolKey
      :icon="playing ? 'mdi-pause' : 'mdi-play'"
      :tone="playing ? 'warn' : 'success'"
      :title="playing ? t('file.pause') : t('file.play')"
      @click="playPause"
    />
    <UiToolKey
      icon="mdi-stop"
      tone="error"
      :title="t('file.stop')"
      :disabled="!running"
      @click="halt"
    />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTuner } from '~/composables/useTuner'
import { useTransport } from '~/composables/useTransport'

const { t } = useI18n()
const { running, starting, start, stop } = useTuner()
const { state, setPaused, seek } = useTransport()

const playing = computed(() => running.value && !state.paused)

async function playPause() {
  if (starting.value) return

  if (!running.value) {
    // Clear a leftover pause first, or Start would run and play nothing.
    if (state.paused) await setPaused(false)
    await start()
    return
  }

  await setPaused(!state.paused)
}

async function halt() {
  await stop()
  await seek(state.from)
}
</script>

<style scoped>
.run {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: 0.25cqw;
}
</style>
