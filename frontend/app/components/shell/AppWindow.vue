<!-- app/components/shell/AppWindow.vue -->
<template>
  <UiDesktopFrame
    class="instrument-frame"
    fill
  >
    <div
      class="window"
      :class="{ 'window--plain': page, 'window--navigation': compact && !landscape }"
    >
      <ShellAppToolbar
        v-if="!page"
        class="window__bar"
        :notes="notes"
      />

      <div class="window__view">
        <TunerView v-show="view === 'tune'" />
        <WorkshopView v-if="view === 'workshop'" />
        <SettingsView v-else-if="view === 'settings'" />
        <ShellHelpView v-else-if="view === 'help'" />
        <ShellAboutView v-else-if="view === 'about'" />
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
import { PAGE_ROOMS, READING_ROOMS, useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'
import { useShortcuts } from '~/composables/useShortcuts'
import { useBackNavigation } from '~/composables/useBackNavigation'
import { windowTitle } from '~/utils/windowTitle'

// Tuner is v-show, not v-if, so leaving and returning does not tear down its four canvases; workshop is v-if.
const { view, setView } = useShell()
const compact = useCompactLayout()
const landscape = useLandscapeLayout()
// A page to read takes the bar's height too where the rooms have a row of their own.
const page = computed(() => compact.value && !landscape.value && PAGE_ROOMS.includes(view.value))
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
  // Calibration is guided tuning (mic open), so the engine stays live there too; and reading the guide
  // does not end a measurement, as its dialog does not in a wide window.
  if (room !== 'tune' && room !== 'calibrate' && !READING_ROOMS.includes(room) && running.value) void stop()
})

// Help and About are rooms only in a compact window; a window grown wide goes back to the tuner.
watch(compact, (now) => {
  if (!now && READING_ROOMS.includes(view.value)) void setView('tune')
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
/* The named rooms share the row's width; Help and About are marks at its end. */
.window__navigation :deep(.key:not(.key--icon)) {
  flex: 1 1 0;
  min-width: 0;
}
.window__navigation :deep(.key__label) { display: inline; }

@media (max-width: 959px), (max-height: 559px) {
  .instrument-frame { --s99ab-frame-inset: 6px; gap: 4px; padding-bottom: 2px; }
  .instrument-frame :deep(.s99ab-frame__status) { height: 22px; margin-top: 0; }
}
</style>
