<template>
  <UiFormField
    :label="t('settings.beatUnitLabel')"
    :hint="t('settings.beatUnitHint')"
  >
    <UiToolGroup
      :model-value="unit"
      :items="unitItems"
      :label="t('settings.beatUnitLabel')"
      @update:model-value="v => unit = v as BeatUnit"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import type { ToolItem } from '@smegg99/s99ab/types'
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

type BeatUnit = 'hz' | 'reference'

const UNITS: readonly BeatUnit[] = ['hz', 'reference']

const { t } = useI18n()
const { config } = useConfigSync()

// The engine reads this too: its beat readouts switch with the next measurement.
const unit = computed<BeatUnit>({
  get: () => (config.tuner?.beat_unit === 'reference' ? 'reference' : 'hz'),
  set: (next: BeatUnit) => {
    config.tuner.beat_unit = next
  },
})

const unitItems = computed<ToolItem[]>(() =>
  UNITS.map(value => ({ value, label: t(`tuner.unit.${value}`) })),
)
</script>
