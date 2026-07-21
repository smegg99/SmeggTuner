<template>
  <div class="d-flex align-center ga-1">
    <UiToolGroup
      :model-value="manual"
      :items="MODES"
      :label="t('tuner.mode.label')"
      @update:model-value="onMode"
    />

    <UiToolKey
      icon="mdi-chevron-left"
      :disabled="!manual"
      :title="t('tuner.step.noteDown')"
      @click="step(-1)"
    />

    <!-- Readout, not a control: fixed width + tabular-nums so the note never shifts the chevrons. -->
    <output
      class="note"
      :class="{ 'note--auto': !manual }"
      :title="t(manual ? 'tuner.mode.manual' : 'tuner.mode.auto')"
    >{{ pinnedName }}</output>

    <UiToolKey
      icon="mdi-chevron-right"
      :disabled="!manual"
      :title="t('tuner.step.noteUp')"
      @click="step(1)"
    />

    <!-- Gate: whether pinning a note sounds a reference tone (opt-in), separate from the output level. -->
    <UiToolKey
      :icon="sounds ? 'mdi-music-note' : 'mdi-music-note-off'"
      :title="t('tuner.sounds.hint')"
      @click="sounds = !sounds"
    />

    <!-- The one output level: governs playback and the reference tone alike; there is no second slider. -->
    <UiLevelControl
      :model-value="volume"
      :on="!muted"
      :active="muted"
      :title="t('tuner.output.volume')"
      :toggle-title="muted ? t('tuner.output.unmute') : t('tuner.output.mute')"
      @toggle="setMuted(!muted)"
      @update:model-value="setVolume"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { AUTO_NOTE, NOTE_MAX, NOTE_MIN, useTuner } from '~/composables/useTuner'
import { useConfigSync } from '~/composables/useConfigSync'
import { useNoteSounds } from '~/composables/useNoteSounds'
import { useOutputLevel } from '~/composables/useOutputLevel'
import { noteName } from '~/utils/tuning'

// Auto: the detector owns the note; manual: the arrows or the note ruler pin it.

// Default pin (A4) when the detector has nothing to hand over.
const DEFAULT_PIN = 69

const { t } = useI18n()
const { config } = useConfigSync()
const { note, manualNote, setManualNote } = useTuner()
const { enabled: sounds } = useNoteSounds()
const { volume, muted, setVolume, setMuted } = useOutputLevel()

const manual = computed(() => manualNote.value !== AUTO_NOTE)

const pinnedName = computed(() => {
  const pinned = manual.value ? manualNote.value : note.value
  if (!pinned) return '--'

  return noteName(pinned, config.tuner?.scale_naming)
})

const MODES = computed<ToolItem[]>(() => [
  { value: false, label: t('tuner.mode.auto') },
  { value: true, label: t('tuner.mode.manual') },
])

// Switching to manual pins whatever the detector is on.
function onMode(next: string | number | boolean) {
  void setManualNote(next ? (note.value || DEFAULT_PIN) : AUTO_NOTE)
}

function step(delta: number) {
  if (!manual.value) return
  void setManualNote(Math.min(NOTE_MAX, Math.max(NOTE_MIN, manualNote.value + delta)))
}
</script>

<style scoped>
.note {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  display: flex;
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 1.75cqh;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  height: 3.9cqh;
  justify-content: center;
  line-height: 1;
  width: 5cqw;
}

.note--auto {
  color: rgb(var(--v-theme-ink2));
  font-weight: 500;
}
</style>
