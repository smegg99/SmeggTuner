<template>
  <div class="window">
    <ShellAppToolbar
      class="window__bar"
      :notes="notes"
    />

    <div class="window__view">
      <TunerView v-show="view === 'tune'" />
      <WorkshopView v-if="view === 'workshop'" />
      <SettingsView v-else-if="view === 'settings'" />
      <CalibrationView v-else-if="view === 'calibrate'" />
    </div>

    <ShellAppStatus class="window__status" />
  </div>
</template>

<script setup lang="ts">
import * as TitleService from '~~bindings/github.com/smegg99/s99wails/services/title/service.js'
import { useAppTitle } from '~/composables/s99wails'
import { useSessionProgress } from '~/composables/useSessionProgress'
import { useShell } from '~/composables/useShell'
import { useShortcuts } from '~/composables/useShortcuts'
import { windowTitle } from '~/utils/windowTitle'

// Tuner is v-show, not v-if, so leaving and returning does not tear down its four canvases; workshop is v-if.
const { view } = useShell()
const { running, stop } = useTuner()
const { table } = useRecord()

// Global shortcuts bound here on the window so they outlive the rooms.
useShortcuts()

// Gates the workshop Print key: nothing to print without readings.
const notes = computed(() => table.value?.rows?.length ?? 0)

// Title is translated here (the Go side prepends the app name and cannot reach vue-i18n); t() reacts to locale changes.
const { t } = useI18n()
const progress = useSessionProgress()
const { setTitle } = useAppTitle(TitleService)

watchEffect(() => {
  setTitle(windowTitle(view.value, progress.value, t))
})

// Leaving the tuning rooms stops the engine; stop() touches only the engine (not the open pass or
// recording), so returning resumes the same pass with nothing lost.
watch(view, (room) => {
  // Calibration is guided tuning (mic open), so the engine stays live there too.
  if (room !== 'tune' && room !== 'calibrate' && running.value) void stop()
})
</script>

<style scoped>
/* container-type is kept for the containment, not for queries: nothing is sized as a fraction of
   the window any more. */
.window {
  background: rgb(var(--v-theme-bg));
  container-type: size;
  display: grid;
  gap: 6.8px;
  grid-template-rows: 42.4px 1fr 43.2px;
  height: 100dvh;
  overflow: hidden;
  /* status bar sits on the bottom edge, so almost no padding below it */
  padding: 6.8px 6.8px 1.2px;
  width: 100vw;
}

.window__view {
  display: grid;
  min-height: 0;
  min-width: 0;
}

.window__view > * {
  grid-area: 1 / 1;
  min-height: 0;
  min-width: 0;
}
</style>
