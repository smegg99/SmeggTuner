<template>
  <UiFormField
    :label="t('settings.detectionLabel')"
    :hint="t(`settings.detection.${detection}Hint`)"
  >
    <UiToolGroup
      :model-value="detection"
      :items="items"
      :label="t('settings.detectionLabel')"
      @update:model-value="v => detection = v as Detection"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import type { ToolItem } from '@smegg99/s99ab/types'
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

type Detection = 'fast' | 'balanced' | 'steady'
const DETECTIONS: readonly Detection[] = ['fast', 'balanced', 'steady']

const { t } = useI18n()
const { config } = useConfigSync()

// The tuner service reads tuner.detection from the config file and reconfigures a running engine.
const detection = computed<Detection>({
  get: () => {
    const v = config.tuner?.detection as Detection | undefined
    return v && DETECTIONS.includes(v) ? v : 'balanced'
  },
  set: (next: Detection) => {
    config.tuner.detection = next
  },
})

const items = computed<ToolItem[]>(() =>
  DETECTIONS.map(value => ({ value, label: t(`settings.detection.${value}`) })),
)
</script>
