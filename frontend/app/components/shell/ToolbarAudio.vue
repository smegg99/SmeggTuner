<template>
  <div class="audio key-row">
    <!-- The app's own output level for file playback and every tone, the note strip's included; the key
         mutes them all and keeps the level. A compact window sets it in Settings. -->
    <UiLevelControl
      v-if="!compact"
      :model-value="volume"
      :on="!muted"
      :active="muted"
      :title="t('tuner.output.volume')"
      :toggle-title="muted ? t('tuner.output.unmute') : t('tuner.output.mute')"
      @toggle="setMuted(!muted)"
      @update:model-value="setVolume"
    />
    <UiToolKey
      icon="mdi-sine-wave"
      :title="t('tuner.signal.open')"
      :active="signalOpen"
      @click="setSignalOpen(!signalOpen)"
    />
  </div>
</template>

<script setup lang="ts">
import { useOutputLevel } from '~/composables/useOutputLevel'
import { useSignal } from '~/composables/useSignal'

const { t } = useI18n()
const compact = useCompactLayout()
const { volume, muted, setVolume, setMuted } = useOutputLevel()
const { open: signalOpen, setOpen: setSignalOpen } = useSignal()
</script>

<style scoped>
.audio { flex: 0 0 auto; }
</style>
