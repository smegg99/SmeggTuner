<template>
  <UiCheckBox
    v-model="toTray"
    :label="t('settings.closeToTray')"
    :hint="t('settings.closeToTrayHint')"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

// close_to_tray decides what the window's X button does; honoured by main.go.
const { t } = useI18n()
const { config } = useConfigSync()

const toTray = computed({
  get: () => config.preferences?.close_to_tray ?? true,
  set: (v: boolean) => { config.preferences.close_to_tray = v },
})
</script>
