<template>
  <UiFormField
    :label="t('settings.humFilterLabel')"
    :hint="running ? t('settings.lockedWhileRunning') : t('settings.humFilterHint')"
  >
    <UiToolGroup
      v-model="hum"
      :items="CHOICES"
      :label="t('settings.humFilterLabel')"
      :disabled="running"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { config } = useConfigSync()
// Locked while running: the notches are applied only at engine start.
const { running } = useTuner()

// One choice of three, not two switches: a mains supply is 50 or 60 Hz or the microphone is not
// hearing it, and off used to be reachable only by turning the lit one back off.
const CHOICES = computed<ToolItem[]>(() => [
  { value: 0, label: t('settings.humFilterNone') },
  { value: 50, label: t('settings.humFilter50') },
  { value: 60, label: t('settings.humFilter60') },
])

// The config keeps a flag per frequency; this is the one place that knows they are alternatives.
// A file with both set reads as 50 and normalises on the next change.
const hum = computed<number>({
  get: () => {
    if (config.audio?.hum_filter_50) return 50
    return config.audio?.hum_filter_60 ? 60 : 0
  },
  set: (hz: number) => {
    config.audio.hum_filter_50 = hz === 50
    config.audio.hum_filter_60 = hz === 60
  },
})
</script>
