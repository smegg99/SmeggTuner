<template>
  <UiFormField
    :label="t('settings.unitLabel')"
    :hint="t('settings.unitHint')"
  >
    <UiToolGroup
      :model-value="unit"
      :items="unitItems"
      :label="t('settings.unitLabel')"
      @update:model-value="v => unit = v as DeviationUnit"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import type { ToolItem } from '@smegg99/s99ab/types'
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { deviationUnit } from '~/utils/deviation'
import type { DeviationUnit } from '~/utils/deviation'

const UNITS: readonly DeviationUnit[] = ['reference', 'cents']

const { t } = useI18n()
const { config } = useConfigSync()

// Display-only: every deviation readout reads this same field, so writing the config moves them
// with no engine call.
const unit = computed<DeviationUnit>({
  get: () => deviationUnit(config.tuner?.deviation_unit),
  set: (next: DeviationUnit) => {
    config.tuner.deviation_unit = next
  },
})

const unitItems = computed<ToolItem[]>(() =>
  UNITS.map(value => ({ value, label: t(`tuner.unit.${value}`) })),
)
</script>
