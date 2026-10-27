<!-- app/components/web/SpeedWarning.vue -->
<template>
  <p
    v-if="perf?.behind"
    class="slow"
    role="alert"
  >
    <v-icon
      icon="mdi-speedometer-slow"
      size="16"
    />
    {{ t('web.slow') }}
  </p>
</template>

<script setup lang="ts">
import { useWebEngine } from '~/web/useWebEngine'

const { t } = useI18n()
const { perf } = useWebEngine()
</script>

<style scoped>
.slow { align-items: center; color: rgb(var(--v-theme-error)); display: flex; font-size: 12px; gap: 6px; margin: 0 0 4px; }
</style>
