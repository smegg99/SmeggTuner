<template>
  <!-- The microphone's run control only; a file's is in FileRun. Icon-only so the key keeps a constant width. -->
  <UiToolKey
    v-if="!file"
    :icon="running ? 'mdi-stop' : 'mdi-play'"
    :tone="running ? 'error' : 'success'"
    :disabled="starting"
    :title="running ? t('tuner.transport.stop') : t('tuner.transport.start')"
    @click="running ? stop() : start()"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTuner } from '~/composables/useTuner'
import { useAudioDevices } from '~/composables/useAudioDevices'

// The engine's state is stated in one place; this is it for the microphone.
const { t } = useI18n()
const { running, starting, start, stop } = useTuner()
const { current } = useAudioDevices()

const file = computed(() => current.value.kind === 'file')
</script>
