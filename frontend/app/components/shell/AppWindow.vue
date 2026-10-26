<!-- app/components/shell/AppWindow.vue -->
<template>
  <UiDesktopFrame
    class="instrument-frame"
    fill
  >
    <div
      class="window"
      :class="{ 'window--plain': compact && !landscape && view === 'settings', 'window--navigation': compact && !landscape }"
    >
      <ShellAppToolbar
        v-if="!compact || landscape || view !== 'settings'"
        class="window__bar"
        :notes="notes"
      />

      <div class="window__view">
        <TunerView v-show="view === 'tune'" />
        <WorkshopView v-if="view === 'workshop'" />
        <SettingsView v-else-if="view === 'settings'" />
        <CalibrationView v-else-if="view === 'calibrate'" />
        <SoundCardView v-else-if="view === 'soundcard'" />
      </div>
      <ShellToolbarViews
        v-if="compact && !landscape"
        class="window__navigation"
      />
    </div>
    <template #status>
      <ShellAppStatus />
    </template>
    <ShellAppTooltip />
  </UiDesktopFrame>
</template>

<script setup lang="ts">
import * as TitleService from '~~bindings/github.com/smegg99/s99wails/v2/services/title/service.js'
import { useAppTitle } from '~/composables/s99wails'
import { useSessionProgress } from '~/composables/useSessionProgress'
import { useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'
import { useShortcuts } from '~/composables/useShortcuts'
import { useBackNavigation } from '~/composables/useBackNavigation'
import { windowTitle } from '~/utils/windowTitle'

// Tuner is v-show, not v-if, so leaving and returning does not tear down its four canvases; workshop is v-if.
const { view } = useShell()
const compact = useCompactLayout()
const landscape = useLandscapeLayout()
const { running, stop } = useTuner()
const { table } = useRecord()

// Global shortcuts bound here on the window so they outlive the rooms.
useShortcuts()
useBackNavigation()

// Gates the workshop Print key: nothing to print without readings.
const notes = computed(() => table.value?.rows?.length ?? 0)

// Title is translated here (the Go side prepends the app name and cannot reach vue-i18n); t() reacts to locale changes.
const { t } = useI18n()
const progress = useSessionProgress()
const { setTitle } = useAppTitle(TitleService)
const { uncalibrated } = useSoundCard()

watchEffect(() => {
  setTitle(windowTitle(view.value, progress.value, t, uncalibrated.value))
})

// Leaving the tuning rooms stops the engine; stop() touches only the engine (not the open pass or
// recording), so returning resumes the same pass with nothing lost.
watch(view, (room) => {
  // Calibration is guided tuning (mic open), so the engine stays live there too.
  if (room !== 'tune' && room !== 'calibrate' && running.value) void stop()
})
</script>

<style scoped>
.window {
  display: grid;
  gap: var(--gap);
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
  min-width: 0;
}

.window--navigation { grid-template-rows: auto minmax(0, 1fr) auto; }
.window--plain { grid-template-rows: minmax(0, 1fr) auto; }

.instrument-frame {
  padding-bottom: 4px;
}

.instrument-frame :deep(.s99ab-frame__status) {
  height: 28px;
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

.window__navigation {
  display: flex;
  gap: var(--s99ab-control-gap);
  min-width: 0;
}
.window__navigation :deep(.key) {
  flex: 1 1 0;
  min-width: 0;
  height: 44px;
}
.window__navigation :deep(.key__label) { display: inline; }

@media (max-width: 959px), (max-height: 559px) {
  .instrument-frame { --s99ab-frame-inset: 6px; gap: 4px; padding-bottom: 2px; }
  .instrument-frame :deep(.s99ab-frame__status) { height: 22px; margin-top: 0; }
}
</style>
