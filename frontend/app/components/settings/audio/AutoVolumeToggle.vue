<template>
  <v-checkbox
    :model-value="autoVolume?.on ?? false"
    :label="t('settings.autoVolume.label')"
    :hint="t(autoVolume && !autoVolume.supported ? 'settings.autoVolume.unsupported' : 'settings.autoVolume.hint')"
    :disabled="!autoVolume?.supported"
    persistent-hint
    @update:model-value="setAutoVolume(!!$event)"
  />
</template>

<script setup lang="ts">
import { useSoundCard } from '~/composables/useSoundCard'

// Per device, on by default: the tuner steers the OS input level while it captures and gives it back at stop.
const { t } = useI18n()
const { autoVolume, setAutoVolume } = useSoundCard()
</script>
