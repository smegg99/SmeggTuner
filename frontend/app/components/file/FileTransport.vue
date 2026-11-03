<!-- app/components/file/FileTransport.vue -->
<template>
  <span class="keys">
    <v-btn
      variant="text"
      :title="t('tuner.file.toStart')"
      :aria-label="t('tuner.file.toStart')"
      class="tone-key"
      style="--key-tone: rgb(var(--v-theme-accent))"
      :icon="'mdi-skip-previous'"
      :data-icon="'mdi-skip-previous'"
      @click="seek(state.from)"
    />
    <v-btn
      variant="text"
      :title="t('tuner.file.toEnd')"
      :aria-label="t('tuner.file.toEnd')"
      class="tone-key"
      style="--key-tone: rgb(var(--v-theme-accent))"
      :icon="'mdi-skip-next'"
      :data-icon="'mdi-skip-next'"
      @click="seek(state.to)"
    />
    <UiToolKey
      icon="mdi-repeat"
      :title="t('tuner.file.loop')"
      :active="state.loop"
      @click="setLoop(!state.loop)"
    />
  </span>

  <template v-if="withVolume">
    <span class="group-rule" />

    <span class="keys">
      <UiLevelControl
        :model-value="volume"
        :on="!muted"
        :active="muted"
        :title="t('tuner.file.volume')"
        :toggle-title="muted ? t('tuner.file.unmute') : t('tuner.file.mute')"
        @toggle="setMuted(!muted)"
        @update:model-value="setVolume"
      />
    </span>
  </template>

  <span class="group-rule" />

  <span class="keys">
    <UiToolKey
      icon="mdi-magnify-minus-outline"
      :title="t('tuner.file.zoomOut')"
      @click="zoom(2, centre)"
    />
    <UiToolKey
      icon="mdi-magnify-plus-outline"
      :title="t('tuner.file.zoomIn')"
      @click="zoom(0.5, centre)"
    />
    <UiToolKey
      icon="mdi-arrow-expand-horizontal"
      :title="t('tuner.file.fit')"
      @click="fit()"
    />
    <UiToolKey
      icon="mdi-selection-remove"
      :title="t('tuner.file.clearSelection')"
      :disabled="!hasSelection"
      @click="clearSelection()"
    />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTransport } from '~/composables/useTransport'
import { useOutputLevel } from '~/composables/useOutputLevel'

// Play/pause/stop live in FileRun; volume is the shared output level (useOutputLevel).
defineProps<{
  /** Carry output level here when the room's toolbar has no volume key of its own. */
  withVolume?: boolean
}>()

const { t } = useI18n()
const {
  state, view, hasSelection,
  seek, setLoop, clearSelection,
  zoom, fit,
} = useTransport()
const { volume, muted, setVolume, setMuted } = useOutputLevel()

// Button zoom has no pointer to anchor on, so it holds the window's middle still.
const centre = computed(() => (view.from + view.to) / 2)
</script>

<style scoped>
/* Plain rows, not UiToolGroup: that is a slotless segmented control and renders empty here. */
.keys {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: var(--s99ab-control-gap);
}
</style>
