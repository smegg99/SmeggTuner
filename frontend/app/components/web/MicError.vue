<!-- app/components/web/MicError.vue -->
<template>
  <p
    v-if="message"
    class="mic"
    role="alert"
  >
    <v-icon
      icon="mdi-microphone-off"
      size="16"
    />
    {{ t(message) }}
  </p>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useWebEngine } from '~/web/useWebEngine'

const { t } = useI18n()
const { micError } = useWebEngine()
const { error } = useAudioDevices()
const message = computed(() => error.value || (micError.value ? `web.mic.${micError.value}` : ''))
</script>

<style scoped>
.mic { align-items: center; color: rgb(var(--v-theme-error)); display: flex; font-size: 12px; gap: 6px; margin: 0 0 4px; }
</style>
