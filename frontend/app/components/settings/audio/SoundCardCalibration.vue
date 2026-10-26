<template>
  <!-- The stored correction this device's readings run with is in the status line; there is nothing to type. -->
  <UiFormField
    :label="t('settings.calibration.label')"
    :hint="t('settings.calibration.hint')"
    :error="error ? t(error) : undefined"
  >
    <div class="cal">
      <span
        class="cal__status"
        :style="{ color: `rgb(var(--v-theme-${tone}))` }"
      >{{ line ? t(line.key, line.params) : '' }}</span>

      <UiToolKey
        v-if="running"
        :label="t('settings.calibration.cancel')"
        :title="t('settings.calibration.cancel')"
        @click="cancelCalibration"
      />
      <UiToolKey
        v-else
        :label="t('settings.calibration.start')"
        :title="t('settings.calibration.startHint')"
        :disabled="!calibration"
        @click="calibrate"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'
import { calibrationLine, calibrationTone } from '~/utils/soundCard'

const { t } = useI18n()
const { calibration, error, startCalibration, cancelCalibration } = useSoundCard()

const { setView } = useShell()

const running = computed(() => calibration.value?.running ?? false)

// The calibration has its own room, which shows what is being measured while it runs.
async function calibrate() {
  if (await setView('soundcard')) await startCalibration()
}
const line = computed(() => (calibration.value ? calibrationLine(calibration.value) : null))
const tone = computed(() => calibrationTone(calibration.value?.status ?? 'uncalibrated'))
</script>

<style scoped>
.cal {
  align-items: center;
  display: flex;
  gap: var(--gap);
  min-width: 0;
}

.cal__status {
  flex: 1 1 auto;
  font-size: 11.5px;
  font-weight: 600;
  min-width: 0;
}
</style>
