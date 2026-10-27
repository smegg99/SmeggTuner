<!-- app/components/calibration/CalibrationNeedsSession.vue -->
<template>
  <!-- Calibration records onto a session's instrument, so it needs one open: the way to the workshop. -->
  <div class="cal__empty">
    <p class="cal__lead">
      {{ t('calibrate.needsSession') }}
    </p>
    <UiToolKey
      icon="mdi-folder-music-outline"
      :label="t('routes.workshop')"
      :title="t('routes.workshop')"
      @click="setView('workshop')"
    />
  </div>
</template>

<script setup lang="ts">
import { useShell } from '~/composables/useShell'

const { t } = useI18n()
const { setView } = useShell()
</script>

<style scoped>
.cal__empty {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: var(--gap-wide);
  height: 100%;
  justify-content: center;
  min-height: 0;
  padding: var(--gap-wide);
  text-align: center;
}

.cal__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 14px;
  margin: 0;
}
</style>
