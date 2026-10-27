<!-- app/components/web/MicError.vue -->
<template>
  <p
    v-if="micError"
    class="mic"
    role="alert"
  >
    <v-icon
      icon="mdi-microphone-off"
      size="16"
    />
    {{ t(`web.mic.${micError}`) }}
  </p>
</template>

<script setup lang="ts">
import { useWebEngine } from '~/web/useWebEngine'

const { t } = useI18n()
const { micError } = useWebEngine()
</script>

<style scoped>
.mic { align-items: center; color: rgb(var(--v-theme-error)); display: flex; font-size: 12px; gap: 6px; margin: 0 0 4px; }
</style>
