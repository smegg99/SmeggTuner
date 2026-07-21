<template>
  <UiFormField
    :label="t('settings.unitLabel')"
    :hint="t('settings.unitHint')"
  >
    <UiToolGroup
      :model-value="unit"
      :items="unitItems"
      :label="t('settings.unitLabel')"
      @update:model-value="v => unit = v as TuningUnit"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import type { TuningUnit } from '~/composables/useTuner'

const UNITS: readonly TuningUnit[] = ['cent', 'hz']

const { t } = useI18n()
const { config } = useConfigSync()

// Display-only: useTuner reads this same field, so writing the config moves the readout with no
// engine call.
const unit = computed<TuningUnit>({
  get: () => (config.tuner?.unit === 'hz' ? 'hz' : 'cent'),
  set: (next: TuningUnit) => {
    config.tuner.unit = next
  },
})

const unitItems = computed<ToolItem[]>(() =>
  UNITS.map(value => ({ value, label: t(`tuner.unit.${value}`) })),
)
</script>
