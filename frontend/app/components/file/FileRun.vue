<template>
  <span class="run">
    <v-btn
      variant="text"
      :title="playing ? t('file.pause') : t('file.play')"
      :aria-label="playing ? t('file.pause') : t('file.play')"
      class="tone-key"
      :style="{ '--key-tone': `rgb(var(--v-theme-${playing ? 'warn' : 'goal'}))` }"
      :icon="playing ? 'mdi-pause' : 'mdi-play'"
      :data-icon="playing ? 'mdi-pause' : 'mdi-play'"
      @click="playPause"
    />
    <v-btn
      variant="text"
      :title="t('file.stop')"
      :disabled="!running"
      :aria-label="t('file.stop')"
      class="tone-key"
      style="--key-tone: rgb(var(--v-theme-reed))"
      :icon="'mdi-stop'"
      :data-icon="'mdi-stop'"
      @click="halt"
    />
  </span>
</template>

<script setup lang="ts">
import { useFileRun } from '~/composables/useFileRun'

const { t } = useI18n()
const { playing, running, playPause, halt } = useFileRun()
</script>

<style scoped>
.run {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: var(--s99ab-control-gap);
}
</style>
