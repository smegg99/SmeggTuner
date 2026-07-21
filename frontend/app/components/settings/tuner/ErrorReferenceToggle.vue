<template>
  <UiFormField
    :label="t('settings.errorReference')"
    :hint="t('settings.errorReferenceHint')"
  >
    <UiToolGroup
      :model-value="reference"
      :items="ITEMS"
      :label="t('settings.errorReference')"
      @update:model-value="v => reference = v as ErrorReference"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { DEFAULT_ERROR_REFERENCE, ERROR_REFERENCE } from '~/types/config'
import type { ErrorReference } from '~/types/config'

// scale/goal are one measurement shown two ways (core/target computes both); this only picks which.
const { t } = useI18n()
const { config } = useConfigSync()

const ITEMS = computed<ToolItem[]>(() => [
  { value: ERROR_REFERENCE.SCALE, label: t('settings.errorReferenceScale') },
  { value: ERROR_REFERENCE.GOAL, label: t('settings.errorReferenceGoal') },
])

const reference = computed<ErrorReference>({
  get: () => config.tuner?.error_reference ?? DEFAULT_ERROR_REFERENCE,
  set: (v: ErrorReference) => { config.tuner.error_reference = v },
})
</script>
