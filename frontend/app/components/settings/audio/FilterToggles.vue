<template>
  <UiFormField
    :label="t('settings.humFilterLabel')"
    :hint="running ? t('settings.lockedWhileRunning') : t('settings.humFilterHint')"
  >
    <div class="d-flex ga-1">
      <UiToolKey
        :label="t('settings.humFilter50')"
        :active="hum50"
        :disabled="running"
        @click="hum50 = !hum50"
      />
      <UiToolKey
        :label="t('settings.humFilter60')"
        :active="hum60"
        :disabled="running"
        @click="hum60 = !hum60"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { config } = useConfigSync()
// Locked while running: the notches are applied only at engine start.
const { running } = useTuner()

// A running engine picks changes up at its next start. 50 and 60 are mutually exclusive:
// turning one on clears the other.
const hum50 = computed<boolean>({
  get: () => config.audio?.hum_filter_50 ?? false,
  set: (on: boolean) => {
    config.audio.hum_filter_50 = on
    if (on) config.audio.hum_filter_60 = false
  },
})

const hum60 = computed<boolean>({
  get: () => config.audio?.hum_filter_60 ?? false,
  set: (on: boolean) => {
    config.audio.hum_filter_60 = on
    if (on) config.audio.hum_filter_50 = false
  },
})
</script>
