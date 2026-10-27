<template>
  <UiPanel
    class="cal tuner-panel"
    flush
  >
    <!-- Untitled: the main bar names the room and the instrument. -->
    <!-- Calibration records onto a session's instrument, so it needs one open. -->
    <CalibrationNeedsSession v-if="!instrument" />

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
  </UiPanel>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, onUnmounted, watchEffect } from 'vue'
import CalibrationNeedsSession from './CalibrationNeedsSession.vue'
import CalibrationSteps from './CalibrationSteps.vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useCalibration } from '~/composables/useCalibration'
import { useShell } from '~/composables/useShell'
import { useTuner } from '~/composables/useTuner'
import { registerPendingEdit } from '~/composables/pendingEdits'
import { registerBackStep } from '~/composables/useBackNavigation'
import { error, keyOf } from '~/composables/sessionStore'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/backend/services/session/service.js'

// The engine stays live across this view; AppWindow owns its lifecycle.
const { pageMeta } = useShell()
const { running, start } = useTuner()
const { instrument, stage, reset, leave, back } = useCalibration()
registerBackStep(back)
const removePending = registerPendingEdit(async () => {
  try {
    await leave()
    await SessionService.Save()
    return true
  }
  catch (failure) {
    error.value = keyOf(failure)
    return false
  }
})

const { current } = useAudioDevices()
const file = computed(() => current.value.kind === 'file')

// reset() picks the starting stage: keyboard if the instrument's range is unknown, sweep setup if known.
onMounted(() => {
  reset()
  if (!running.value) void start()
})

// Leaving (this view is a v-if) unpins the note and disarms a sweep; AppWindow stops the engine.
onUnmounted(() => {
  removePending()
  void leave().catch((failure) => {
    error.value = keyOf(failure)
  })
})

watchEffect(() => {
  pageMeta.value = instrument.value?.name ?? ''
})
onBeforeUnmount(() => {
  pageMeta.value = ''
})
</script>

<style scoped>
.cal__body {
  display: flex;
  flex-direction: column;
  gap: var(--gap-wide);
  min-height: 0;
  overflow: hidden;
  padding: var(--gap-wide);
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
  gap: var(--gap-wide);
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
  block-size: 168px;
  flex: 0 0 auto;
  min-height: 0;
  text-align: start;
}

/* The window cannot go below 640 tall (backend/internal/app/window.go). Everything above is sized for a
   comfortable one; at a short one the step is trimmed to fit, because a calibration screen you
   have to scroll to reach Capture is worse than a smaller note box. */
@media (max-height: 780px) {
  .cal__body {
    gap: var(--gap);
    padding: var(--gap-wide);
  }

  .cal__flow {
    gap: var(--gap);
  }

  .cal__input {
    block-size: 131px;
  }

  /* Reaches into the step components on purpose: this is the one place that knows how much room
     the whole screen has. */
  .cal__flow :deep(.rng),
  .cal__flow :deep(.sw) {
    gap: var(--gap);
  }

  .cal__flow :deep(.rng) {
    --cr-h: 101.5px;
    --cr-note: 35px;
    --cr-w: 101.5px;
  }

  .cal__flow :deep(.sw) {
    --cr-h: 84px;
    --cr-note: 31.5px;
    --cr-w: 227.5px;
  }
}
@media (max-width: 959px), (max-height: 559px) {
  .cal__body { padding: 6px; gap: 6px; }
  .cal__flow { overflow: hidden; }
  .cal__input { flex: 1 1 90px; max-height: 160px; }
}
@media (max-height: 559px) and (min-width: 560px) {
  .cal__body { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .cal__flow { align-items: stretch; }
  .cal__flow > :first-child, .cal__flow > :last-child { margin: 0; }
  .cal__input { block-size: auto; max-height: none; }
  .cal__flow :deep(.rng), .cal__flow :deep(.sw) { --cr-h: 52px; --cr-note: 25px; --cr-w: 100%; gap: 4px; }
  .cal__flow :deep(.rng__prompt), .cal__flow :deep(.sw__prompt) { font-size: 12px; }
  .cal__flow :deep(.sw__hint) { font-size: 10px; }
}
</style>
