<template>
  <div class="calibration-tools key-row">
    <UiToolKey
      icon="mdi-arrow-left"
      :label="t('calibrate.leave')"
      :title="t('calibrate.leave')"
      @click="setView('workshop')"
    />

    <!-- Same source selection as the tune room, so a recording picked here is what the engine reads there. -->
    <ShellToolbarSource />
  </div>
</template>

<script setup lang="ts">
import { useShell } from '~/composables/useShell'

const { t } = useI18n()
const { setView } = useShell()
</script>

<style scoped>
.calibration-tools { width: 100%; min-width: 0; }
@media (max-width: 559px) {
  .calibration-tools > .key :deep(.key__label) { display: none; }
}
</style>
