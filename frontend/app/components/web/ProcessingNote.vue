<!-- app/components/web/ProcessingNote.vue -->
<template>
  <p
    v-if="micOpen && processing.length"
    class="note"
    role="status"
  >
    <v-icon
      icon="mdi-alert-outline"
      size="16"
    />
    {{ t('web.processingOn', { list: processing.map(p => t(`web.processing.${p}`)).join(', ') }) }}
  </p>
</template>

<script setup lang="ts">
import { useWebEngine } from '~/web/useWebEngine'

const { t } = useI18n()
const { processing, micOpen } = useWebEngine()
</script>

<style scoped>
.note { align-items: center; color: rgb(var(--v-theme-warning)); display: flex; font-size: 12px; gap: 6px; margin: 0 0 4px; }
</style>
