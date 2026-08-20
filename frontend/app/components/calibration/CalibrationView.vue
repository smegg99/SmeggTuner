<template>
  <UiPanelCard class="cal">
    <template #head>
      <span class="panel__title">{{ t('calibrate.title') }}</span>
      <span class="cal__sp" />
      <span class="panel__meta">{{ instrumentName }}</span>
    </template>

    <!-- Calibration records onto a session's instrument, so it needs one open. -->
    <div
      v-if="!instrument"
      class="cal__empty"
    >
      <p class="cal__lead">
        {{ t('calibrate.needsSession') }}
      </p>
      <UiToolKey
        icon="mdi-folder-music-outline"
        :label="t('routes.workshop')"
        @click="setView('workshop')"
      />
    </div>

    <div
      v-else
      class="cal__body"
    >
      <div class="cal__flow">
        <!-- Keyboard first (stage range), because the sweep walks the range it finds. -->
        <CalibrationSteps :stage="stage" />

        <CalibrationRange v-if="stage === 'range'" />
        <CalibrationSweep v-else />
      </div>

      <!-- File selection uses FilePanel (with-volume since this room's toolbar has no volume key); mic uses the input scope. -->
      <FilePanel
        v-if="file"
        with-volume
        class="cal__input"
      />
      <TunerInputPanel
        v-else
        class="cal__input"
      />
    </div>
  </UiPanelCard>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import CalibrationSteps from './CalibrationSteps.vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useCalibration } from '~/composables/useCalibration'
import { useShell } from '~/composables/useShell'
import { useTuner } from '~/composables/useTuner'

// The engine stays live across this view; AppWindow owns its lifecycle.
const { t } = useI18n()
const { setView } = useShell()
const { running, start } = useTuner()
const { instrument, stage, reset, leave } = useCalibration()

const { current } = useAudioDevices()
const file = computed(() => current.value.kind === 'file')

// reset() picks the starting stage: keyboard if the instrument's range is unknown, sweep setup if known.
onMounted(() => {
  reset()
  if (!running.value) void start()
})

// Leaving (this view is a v-if) unpins the note and disarms a sweep; AppWindow stops the engine.
onUnmounted(() => {
  void leave()
})

const instrumentName = computed(() => {
  const i = instrument.value
  return i?.name ?? ''
})
</script>

<style scoped>
.cal__sp {
  flex: 1 1 auto;
}

.cal__empty {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  justify-content: center;
  min-height: 0;
  padding: 24px 22px;
  text-align: center;
}

.cal__body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
  overflow: hidden;
  padding: 24px 22px;
}

/* Auto margins on the ends rather than justify-content: center. Centred content that outgrows the
   pane overflows in BOTH directions, and the half above the top is unreachable - at the smallest
   window that put the step circles behind the panel head and Capture on top of the input scope.
   Auto margins collapse once the content is taller, so it degrades to scrolling from the top
   rather than to overlap. It should not come to that: the rule at the bottom of this file trims
   the step to fit the shortest window the app allows. */
.cal__flow {
  align-items: center;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
  overflow-y: auto;
  text-align: center;
}

.cal__flow > :first-child {
  margin-block-start: auto;
}

.cal__flow > :last-child {
  margin-block-end: auto;
}

.cal__input {
  block-size: 192px;
  flex: 0 0 auto;
  min-height: 0;
  text-align: start;
}

.cal__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 16px;
  margin: 0;
}

/* The window cannot go below 640 tall (internal/app/window.go). Everything above is sized for a
   comfortable one; at a short one the step is trimmed to fit, because a calibration screen you
   have to scroll to reach Capture is worse than a smaller note box. */
@media (max-height: 780px) {
  .cal__body {
    gap: 10px;
    padding: 12px 22px;
  }

  .cal__flow {
    gap: 10px;
  }

  .cal__input {
    block-size: 150px;
  }

  /* Reaches into the step components on purpose: this is the one place that knows how much room
     the whole screen has. */
  .cal__flow :deep(.rng),
  .cal__flow :deep(.sw) {
    gap: 10px;
  }

  .cal__flow :deep(.rng) {
    --cr-h: 116px;
    --cr-note: 40px;
    --cr-w: 116px;
  }

  .cal__flow :deep(.sw) {
    --cr-h: 96px;
    --cr-note: 36px;
    --cr-w: 260px;
  }
}
</style>
