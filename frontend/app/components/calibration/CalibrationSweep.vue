<template>
  <CalibrationSweepSetup
    v-if="stage === 'setup'"
    :registers="registers"
    :bass-reeds="instrument?.bassReeds"
    :bass-registers="instrument?.bassRegisters"
    @begin="begin"
    @redo-range="redoRange"
  />

  <div
    v-else-if="stage === 'sweep'"
    class="sw"
  >
    <div class="sw__progress">
      <div class="sw__bar">
        <div
          class="sw__bar-fill"
          :style="{ width: `${pct}%` }"
        />
      </div>
      <span class="sw__count">{{ t('calibrate.sweep.progress', { done: progress.done, total: progress.total }) }}</span>
    </div>

    <p
      v-if="currentStep"
      class="sw__ask"
    >
      <span class="sw__asknote">{{ stepLabel(currentStep) }}</span>
    </p>

    <CalibrationReadout
      :note="readoutNote ? noteName(readoutNote) : '--'"
      :sure="!!heard"
      :status="status"
      :progress="lockProgress"
      box-width="30cqw"
      box-height="18cqh"
      note-size="6cqh"
    />

    <div class="sw__keys">
      <UiToolKey
        icon="mdi-debug-step-over"
        :label="t('calibrate.sweep.skip')"
        @click="skip"
      />
      <UiToolKey
        icon="mdi-stop"
        tone="error"
        :label="t('calibrate.sweep.stop')"
        @click="finish"
      />
    </div>
  </div>

  <div
    v-else
    class="sw sw--done"
  >
    <v-icon
      icon="mdi-check-circle-outline"
      class="sw__done-icon"
    />
    <p class="sw__prompt">
      {{ t('calibrate.sweep.done', capturedCount) }}
    </p>
    <div class="sw__keys">
      <UiToolKey
        icon="mdi-table"
        :label="t('calibrate.sweep.review')"
        @click="setView('workshop')"
      />
      <UiToolKey
        icon="mdi-refresh"
        :label="t('calibrate.sweep.again')"
        @click="reset"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import CalibrationReadout from './CalibrationReadout.vue'
import CalibrationSweepSetup from './CalibrationSweepSetup.vue'
import { useCalibration } from '~/composables/useCalibration'
import { useShell } from '~/composables/useShell'
import { useConfigSync } from '~/composables/useConfigSync'
import { noteName as toNoteName, pitchClassName } from '~/utils/tuning'
import type { Step } from '~/composables/calibrationSteps'

const { t } = useI18n()
const { setView } = useShell()
const { config } = useConfigSync()
const {
  stage, registers, instrument, currentStep, heard, shown, liveNote, status, lockProgress,
  captured, capturedCount, progress, begin, skip, finish, redoRange, reset,
} = useCalibration()

const pct = computed(() => (progress.value.total ? (progress.value.done / progress.value.total) * 100 : 0))

const readoutNote = computed(() => heard.value || liveNote.value || shown.value)

function noteName(n: number): string {
  return toNoteName(n, config.tuner?.scale_naming)
}

// A bass step names a button, not a key: the pitch class alone, whatever octave the ladder answers at.
function stepLabel(s: Step): string {
  return s.pc !== undefined ? pitchClassName(s.pc, config.tuner?.scale_naming) : noteName(s.note)
}

watch(captured, (got) => {
  if (got && stage.value === 'sweep') skip()
})
</script>

<style scoped>
.sw {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 2cqh;
  justify-content: center;
  width: 100%;
}

.sw__prompt {
  color: rgb(var(--v-theme-ink));
  font-size: 2.2cqh;
  margin: 0;
  max-width: 44ch;
}

.sw__progress {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 0.6cqh;
  width: 44cqw;
}

.sw__bar {
  background: rgb(var(--v-theme-well));
  border-radius: 999px;
  height: 0.9cqh;
  overflow: hidden;
  width: 100%;
}

.sw__bar-fill {
  background: rgb(var(--v-theme-ink2));
  height: 100%;
  transition: width 160ms;
}

.sw__count {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  font-variant-numeric: tabular-nums;
}

.sw__ask {
  align-items: baseline;
  display: flex;
  gap: 1.2cqw;
  margin: 0;
}

.sw__asknote {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 4.5cqh;
  font-weight: 700;
}

.sw__keys {
  display: flex;
  gap: 0.6cqw;
}

.sw__done-icon {
  color: rgb(var(--v-theme-goal));
  font-size: 8cqh;
}
</style>
