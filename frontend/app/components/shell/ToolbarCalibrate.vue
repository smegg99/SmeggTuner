<template>
  <div class="d-flex align-center ga-1">
    <UiToolKey
      icon="mdi-arrow-left"
      :label="t('calibrate.leave')"
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
