<template>
  <v-checkbox
    v-model="stopAfterLock"
    :label="t('settings.stopAfterLock')"
    :hint="t('settings.stopAfterLockHint')"
    persistent-hint
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

// stopAfterLock: freeze the reading when a note locks so it holds still while you file; honoured by services/tuner.
const { t } = useI18n()
const { config } = useConfigSync()

const stopAfterLock = computed({
  get: () => config.tuner?.stop_after_lock ?? false,
  set: (v: boolean) => { config.tuner.stop_after_lock = v },
})
</script>
