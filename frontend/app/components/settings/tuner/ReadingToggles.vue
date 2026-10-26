<template>
  <v-checkbox
    v-model="continuousManual"
    :label="t('settings.continuousManual')"
    :hint="t('settings.continuousManualHint')"
    persistent-hint
  />

  <v-checkbox
    v-model="lockA4"
    :label="t('settings.lockA4')"
    :hint="t('settings.lockA4Hint')"
    persistent-hint
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

// continuousManual: with a note pinned, keep updating even when the engine isn't confident.
// lockA4: services/tuner refuses every A4 change while it is on, with or without a session.
const { t } = useI18n()
const { config } = useConfigSync()

const continuousManual = computed({
  get: () => config.tuner?.continuous_update_manual ?? false,
  set: (v: boolean) => { config.tuner.continuous_update_manual = v },
})

const lockA4 = computed({
  get: () => config.tuner?.lock_a4 ?? false,
  set: (v: boolean) => { config.tuner.lock_a4 = v },
})
</script>
