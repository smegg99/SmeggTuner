<template>
  <div class="note-tools key-row">
    <!-- Lit while the detector owns the note; the note key beside it pins one. -->
    <UiToolKey
      :label="t('tuner.mode.auto')"
      :title="t('tuner.mode.autoHint')"
      :active="!manual"
      :disabled="chords"
      @click="manual && setAuto(true)"
    />

    <div class="note-step">
      <UiToolKey
        icon="mdi-chevron-left"
        :disabled="!manual"
        :title="t('tuner.step.noteDown')"
        @click="stepNote(-1)"
      />

      <!-- The note on screen, as a toggle: pressed it pins that note and stays lit; pressed again the
           detector takes over. Fixed width so the chevrons never move. -->
      <UiToolKey
        class="note"
        :label="pinnedName"
        :active="manual"
        :title="t(manual ? 'tuner.mode.pinned' : 'tuner.mode.pinHint')"
        :disabled="chords"
        @click="setAuto(manual)"
      />

      <UiToolKey
        icon="mdi-chevron-right"
        :disabled="!manual"
        :title="t('tuner.step.noteUp')"
        @click="stepNote(1)"
      />
    </div>

    <!-- Wide windows carry the mode and reed keys here; compact ones give them a row of their own. -->
    <template v-if="!compact && !sessionReeds">
      <span class="group-rule" />
      <TunerModePanel />
      <span class="group-rule" />
    </template>

    <!-- Compact windows carry the output keys in the top bar, beside the input. -->
    <ShellToolbarAudio v-if="!WEB && !compact" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'
import { useConfigSync } from '~/composables/useConfigSync'
import { noteName } from '~/utils/tuning'
import { WEB } from '~/utils/features'

// Auto: the detector owns the note; manual: the arrows or the note ruler pin it. Chords picks its own notes.
const { t } = useI18n()
const { config } = useConfigSync()
const { note, manualNote, mode, sessionReeds, setAuto, stepNote } = useTuner()
const compact = useCompactLayout()

const manual = computed(() => manualNote.value !== AUTO_NOTE)
const chords = computed(() => !sessionReeds.value && mode.value === 'chords')

const pinnedName = computed(() => {
  const pinned = manual.value ? manualNote.value : note.value
  if (!pinned) return '--'

  return noteName(pinned, config.tuner?.scale_naming)
})
</script>

<style scoped>
.note-step { display: flex; align-items: center; gap: var(--s99ab-control-gap); flex: 0 0 auto; }

.note {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  justify-content: center;
  width: 48px;
}
</style>
