<template>
  <UiCheckBox
    v-model="stopAfterLock"
    :label="t('settings.stopAfterLock')"
    :hint="t('settings.stopAfterLockHint')"
  />

  <UiCheckBox
    v-model="continuousManual"
    :label="t('settings.continuousManual')"
    :hint="t('settings.continuousManualHint')"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

// stopAfterLock: freeze the reading when a note locks so it holds still while you file.
// continuousManual: with a note pinned, keep updating even when the engine isn't confident.
// Both are honoured by services/tuner.
const { t } = useI18n()
const { config } = useConfigSync()

const stopAfterLock = computed({
  get: () => config.tuner?.stop_after_lock ?? false,
  set: (v: boolean) => { config.tuner.stop_after_lock = v },
})

const continuousManual = computed({
  get: () => config.tuner?.continuous_update_manual ?? false,
  set: (v: boolean) => { config.tuner.continuous_update_manual = v },
})
</script>
