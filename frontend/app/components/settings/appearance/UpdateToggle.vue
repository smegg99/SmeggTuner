<!-- app/components/settings/appearance/UpdateToggle.vue -->
<template>
  <v-checkbox
    v-if="!platform.android"
    v-model="checkUpdates"
    :label="t('settings.checkUpdates')"
    :hint="t('settings.checkUpdatesHint')"
    persistent-hint
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDisplay } from 'vuetify'
import { useConfigSync } from '~/composables/useConfigSync'
import { updateActions } from '~/composables/s99wails'

const { t } = useI18n()
const { platform } = useDisplay()
const { config, saveConfigNow } = useConfigSync()

async function setCheckUpdates(value: boolean) {
  config.preferences.check_updates = value
  await saveConfigNow()
  if (value && config.preferences.check_updates) await updateActions.check().catch(() => undefined)
}

const checkUpdates = computed({
  get: () => config.preferences?.check_updates ?? true,
  set: (value: boolean) => { void setCheckUpdates(value) },
})
</script>
