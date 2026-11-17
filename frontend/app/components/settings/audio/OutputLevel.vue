<!-- app/components/settings/audio/OutputLevel.vue -->
<template>
  <!-- The app's own output level for file playback and every tone. A wide window carries it in the bar
       (ShellToolbarAudio); a compact one sets it here. -->
  <UiFormField :label="t('tuner.output.volume')">
    <UiLevelControl
      class="output-level"
      :model-value="volume"
      :on="!muted"
      :active="muted"
      :title="t('tuner.output.volume')"
      :toggle-title="muted ? t('tuner.output.unmute') : t('tuner.output.mute')"
      @toggle="setMuted(!muted)"
      @update:model-value="setVolume"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { useOutputLevel } from '~/composables/useOutputLevel'

const { t } = useI18n()
const { volume, muted, setVolume, setMuted } = useOutputLevel()
</script>

<style scoped>
/* The slider takes the field's width. */
.output-level { display: flex; }
.output-level :deep(.level-control__slider) { flex: 1 1 auto; }
.output-level :deep(.tuner-volume) { flex: 1 1 auto; width: auto; }
</style>
