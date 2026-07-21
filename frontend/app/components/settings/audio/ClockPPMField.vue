<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.clockPpmLabel')"
    :hint="running ? t('settings.lockedWhileRunning') : t('settings.clockPpmHint')"
  >
    <UiNumberInput
      :id="id"
      v-model="draft"
      :suffix="t('settings.clockPpmUnit')"
      :disabled="running"
      @commit="commit"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { useConfigSync } from '~/composables/useConfigSync'
import { useDraftNumber } from '~/composables/useDraftNumber'
import { useTuner } from '~/composables/useTuner'

// core/dsp corrects the sample-clock error; the range only guards a typo from moving the scale.
const PPM_MIN = -1000
const PPM_MAX = 1000

const { t } = useI18n()
const { config } = useConfigSync()
// Locked while running: the correction is applied only at engine start.
const { running } = useTuner()

const { draft, commit } = useDraftNumber(
  () => config.audio?.clock_ppm ?? 0,
  value => (config.audio.clock_ppm = value),
  { min: PPM_MIN, max: PPM_MAX, step: 0.1 },
)
</script>
