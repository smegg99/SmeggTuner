<template>
  <div class="bar">
    <template v-if="view === 'tune'">
      <ShellToolbarTransport />
      <ShellToolbarRecord />

      <div class="bar__divider" />

      <ShellToolbarSource />

      <div class="bar__divider" />

      <ShellToolbarNote />

      <div class="bar__divider" />

      <ShellToolbarReading />
    </template>

    <!-- v-else-if, not v-else: a bare v-else would also catch the settings room, which has no toolbar keys. -->
    <ShellToolbarCalibrate v-else-if="view === 'calibrate'" />

    <ShellToolbarWorkshop
      v-else-if="view === 'workshop'"
      :notes="notes"
    />

    <div class="bar__spacer" />

    <ShellToolbarViews />
  </div>
</template>

<script setup lang="ts">
import { useShell } from '~/composables/useShell'

defineProps<{
  /** notes on screen; enables the workshop Print key */
  notes?: number
}>()

const { view } = useShell()
</script>

<style scoped>
.bar {
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 0.45cqh;
  display: flex;
  gap: 0.5cqw;
  min-width: 0;
  overflow: hidden;
  padding: 0 0.65cqw;
}

.bar__divider {
  background: rgb(var(--v-theme-line));
  flex: 0 0 1px;
  height: 2.5cqh;
  margin: 0 0.35cqw;
}

.bar__spacer {
  flex: 1 1 auto;
}
</style>
