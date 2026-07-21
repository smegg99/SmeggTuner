<template>
  <span class="keys">
    <UiToolKey
      icon="mdi-skip-previous"
      tone="accent"
      :title="t('file.toStart')"
      @click="seek(state.from)"
    />
    <UiToolKey
      icon="mdi-skip-next"
      tone="accent"
      :title="t('file.toEnd')"
      @click="seek(state.to)"
    />
    <UiToolKey
      icon="mdi-repeat"
      :title="t('file.loop')"
      :active="state.loop"
      @click="setLoop(!state.loop)"
    />
  </span>

  <template v-if="withVolume">
    <span class="rule" />

    <span class="keys">
      <UiLevelControl
        :model-value="volume"
        :on="!muted"
        :active="muted"
        :title="t('file.volume')"
        :toggle-title="muted ? t('file.unmute') : t('file.mute')"
        @toggle="setMuted(!muted)"
        @update:model-value="setVolume"
      />
    </span>
  </template>

  <span class="rule" />

  <span class="keys">
    <UiToolKey
      icon="mdi-magnify-minus-outline"
      :title="t('file.zoomOut')"
      @click="zoom(2, centre)"
    />
    <UiToolKey
      icon="mdi-magnify-plus-outline"
      :title="t('file.zoomIn')"
      @click="zoom(0.5, centre)"
    />
    <UiToolKey
      icon="mdi-arrow-expand-horizontal"
      :title="t('file.fit')"
      @click="fit()"
    />
    <UiToolKey
      icon="mdi-selection-remove"
      :title="t('file.clearSelection')"
      :disabled="!hasSelection"
      @click="clearSelection()"
    />
  </span>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useTuner } from '~/composables/useTuner'
import { useTransport } from '~/composables/useTransport'
import { useOutputLevel } from '~/composables/useOutputLevel'

// Play/pause/stop live in FileRun; volume is the shared output level (useOutputLevel).
defineProps<{
  /** Carry output level here when the room's toolbar has no volume key of its own. */
  withVolume?: boolean
}>()

const { t } = useI18n()
const { running, starting, start } = useTuner()
const {
  state, view, hasSelection,
  seek, setPaused, setLoop, clearSelection,
  zoom, fit, fitSelection,
} = useTransport()
const { volume, muted, setVolume, setMuted } = useOutputLevel()

/** Same as the play key in FileRun. */
async function playPause() {
  if (starting.value) return

  if (!running.value) {
    if (state.paused) await setPaused(false)
    await start()
    return
  }
  await setPaused(!state.paused)
}

// Button zoom has no pointer to anchor on, so it holds the window's middle still.
const centre = computed(() => (view.from + view.to) / 2)

function onKey(event: KeyboardEvent) {
  if (!state.available) return

  const el = event.target as HTMLElement | null
  if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) return

  const keys: Record<string, () => void> = {
    // Space must start the engine if it is down, not just toggle a flag on a dead transport.
    ' ': () => void playPause(),
    'Home': () => void seek(state.from),
    'End': () => void seek(state.to),
    'l': () => void setLoop(!state.loop),
    'm': () => void setMuted(!muted.value),
    'f': () => fit(),
    's': () => fitSelection(),
    '+': () => zoom(0.5, centre.value),
    '=': () => zoom(0.5, centre.value),
    '-': () => zoom(2, centre.value),
  }

  const run = keys[event.key.length === 1 ? event.key.toLowerCase() : event.key]
  if (!run) return

  event.preventDefault() // space would otherwise scroll the page
  run()
}

onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<style scoped>
/* Plain rows, not UiToolGroup: that is a slotless segmented control and renders empty here. */
.keys {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: 0.3cqw;
}

.rule {
  background: rgb(var(--v-theme-line));
  block-size: 2.4cqh;
  flex: 0 0 1px;
  margin-inline: 0.45cqw;
}
</style>
