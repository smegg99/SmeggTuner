<!-- app/components/web/AppWindow.vue -->
<!-- The web edition's window: the desktop window's frame, toolbar and status bar, with only the tune and settings rooms. -->
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
        :notes="0"
      />

      <div class="window__view">
        <div
          v-show="view === 'tune'"
          class="window__tune"
        >
          <WebMicError />
          <WebSpeedWarning />
          <WebProcessingNote />
          <TunerView />
        </div>
        <SettingsView v-if="view === 'settings'" />
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
    <WebNotice />
  </UiDesktopFrame>
</template>

<script setup lang="ts">
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useConfigSync } from '~/composables/useConfigSync'
import { initializeThemeSync } from '~/composables/useThemeSync'
import { useShell } from '~/composables/useShell'
import { useWebEngine } from '~/web/useWebEngine'
import { useWakeLock } from '~/web/wakeLock'

useConfigSync()
initializeThemeSync()

const { view, setView } = useShell()
const compact = useCompactLayout()
const landscape = useLandscapeLayout()
const { running, stop } = useTuner()
const { micOpen, openMic, closeMic } = useWebEngine()
const { refresh: refreshDevices } = useAudioDevices()

// Only two rooms exist here; anything else the shared toolbar might ask for lands back on the tuner.
watch(view, (room) => {
  if (room !== 'tune' && room !== 'settings') void setView('tune')
  if (room !== 'tune' && running.value) void stop()
}, { immediate: true })

// The microphone is open exactly while the engine runs, so the browser's recording indicator means what it says;
// without one the run stops again and WebMicError says why.
watch(running, async (on) => {
  if (on && !micOpen.value && !(await openMic())) void stop()
  if (!on && micOpen.value) closeMic()
})

// The browser names its microphones only once one has been opened, so the list is read again then.
watch(micOpen, (open) => {
  if (open) void refreshDevices()
})

useWakeLock(running)
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
.instrument-frame { padding-bottom: 4px; }
.instrument-frame :deep(.s99ab-frame__status) { height: 28px; }
.window__view { display: grid; min-height: 0; min-width: 0; }
.window__view > * { grid-area: 1 / 1; min-height: 0; min-width: 0; }
.window__tune { display: grid; grid-template-rows: auto auto auto minmax(0, 1fr); min-height: 0; min-width: 0; }
.window__navigation { display: flex; gap: var(--s99ab-control-gap); min-width: 0; }
.window__navigation :deep(.key) { flex: 1 1 0; min-width: 0; height: 44px; }
.window__navigation :deep(.key__label) { display: inline; }
@media (max-width: 959px), (max-height: 559px) {
  .instrument-frame { --s99ab-frame-inset: 6px; gap: 4px; padding-bottom: 2px; }
  .instrument-frame :deep(.s99ab-frame__status) { height: 22px; margin-top: 0; }
}
</style>
