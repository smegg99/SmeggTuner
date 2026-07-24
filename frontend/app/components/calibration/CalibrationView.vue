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
  gap: 2cqh;
  height: 100%;
  justify-content: center;
  min-height: 0;
  padding: 3cqh 2cqw;
  text-align: center;
}

.cal__body {
  display: flex;
  flex-direction: column;
  gap: 2cqh;
  min-height: 0;
  padding: 3cqh 2cqw;
}

.cal__flow {
  align-items: center;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 2cqh;
  justify-content: center;
  min-height: 0;
  text-align: center;
}

.cal__input {
  block-size: 24cqh;
  flex: 0 0 auto;
  min-height: 0;
  text-align: start;
}

.cal__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 2cqh;
  margin: 0;
}
</style>
