<template>
  <v-checkbox
    v-if="!platform.android"
    v-model="toTray"
    :label="t('settings.closeToTray')"
    :hint="t('settings.closeToTrayHint')"
    persistent-hint
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDisplay } from 'vuetify'
import { useConfigSync } from '~/composables/useConfigSync'

// close_to_tray decides what the window's X button does; honoured by backend/internal/app/tray.go.
const { t } = useI18n()
const { platform } = useDisplay()
const { config } = useConfigSync()

const toTray = computed({
  get: () => config.preferences?.close_to_tray ?? true,
  set: (v: boolean) => { config.preferences.close_to_tray = v },
})
</script>
