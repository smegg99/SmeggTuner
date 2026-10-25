<!-- app/components/tuner/NotePanel.vue -->
<template>
  <div
    class="note"
    :class="{ 'note--quiet': shownState === 'initializing', 'note--locked': locked }"
  >
    <!-- The detected note is the card: it takes every pixel the rows under it leave. -->
    <div
      class="note__glyph"
      :class="{ 'note__glyph--chord': !!label }"
    >
      <span v-if="label">{{ label }}</span>
      <span
        v-else
        class="note__pitch"
      ><span>{{ pitchClass || '--' }}</span><i>{{ octave }}</i></span>
    </div>

    <span class="note__hz">
      <span>{{ frequency }}</span><span>Hz</span>
      <span
        v-if="transpose > 0"
        class="note__transpose"
        :title="t('tuner.reference.transposed', { n: transpose })"
        :aria-label="t('tuner.reference.transposed', { n: transpose })"
      >+{{ transpose }}</span>
    </span>
    <span
      class="note__status"
      :title="shownState === 'tooQuiet' ? t('tuner.state.tooQuiet') : caption"
      aria-live="polite"
    >
      <v-icon
        v-if="locked"
        icon="mdi-lock"
        size="12px"
      />
      {{ caption }}
    </span>
    <TunerA4Stepper class="note__a4" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useReadoutStatus, type ReadoutStatus } from '~/composables/useReadoutStatus'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { noteName, label, lock, running, state, scalePitch, transpose } = useTuner()

const rawStatus = computed<ReadoutStatus>(() => {
  if (!running.value) return state.value === 'deviceLost' ? 'deviceLost' : 'stopped'
  if (lock.value.visible) return lock.value.recorded ? 'recorded' : 'locked'
  return state.value === 'tooLoud' ? 'running' : state.value
})
const shownState = useReadoutStatus(rawStatus, useDelayConfig())
const locked = computed(() => shownState.value === 'locked' || shownState.value === 'recorded')
const caption = computed(() => {
  if (shownState.value === 'initializing') return t('tuner.warmup.quiet')
  if (shownState.value === 'tooQuiet') return t('tuner.panel.quiet')
  if (locked.value) return t(shownState.value === 'recorded' ? 'tuner.panel.recorded' : 'tuner.panel.locked')
  return shownState.value === 'stopped' ? '' : t(`tuner.state.${shownState.value}`)
})
const frequency = computed(() => shownState.value !== 'initializing' && scalePitch.value > 0 ? scalePitch.value.toFixed(2) : '--')

// Octave is the trailing digits of the note name, split off to style it small.
const pitchClass = computed(() => noteName.value.replace(/-?\d+$/, ''))
const octave = computed(() => noteName.value.match(/-?\d+$/)?.[0] ?? '')
</script>

<style scoped>
/* Every readout keeps its slot, including silence, warm-up and stopped states. The card is a size
   container, so the note grows with whatever height the window gives the spectrum row. */
.note {
  container-type: size;
  display: grid;
  grid-template-areas: "glyph" "hz" "status" "a4";
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: minmax(0, 1fr) 18px 16px 26px;
  align-items: center;
  row-gap: 3px;
  min-width: 0;
  min-height: 0;
  padding: 4px 6px 6px;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: var(--app-radius-panel);
  text-align: center;
}

.note--locked {
  background-image: repeating-linear-gradient(135deg, rgba(var(--v-theme-ink), 0.045) 0 6px, transparent 6px 12px);
  border-color: rgba(var(--v-theme-ink), 0.35);
}
.note--quiet { background-color: rgba(var(--v-theme-warn), 0.14); border-color: rgb(var(--v-theme-warn)); }
.note--quiet .note__status, .note--quiet .note__glyph { color: rgb(var(--v-theme-warn)); }

/* What the three rows under it leave, as a font size: 100cqh is the card. */
.note__glyph {
  grid-area: glyph;
  display: grid;
  place-items: center;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--v-theme-ink));
  font-size: clamp(26px, calc(100cqh - 82px), 88px);
  font-weight: 500;
  line-height: 1;
}
.note__glyph--chord { font-size: clamp(20px, calc((100cqh - 82px) * 0.6), 44px); }
.note__pitch { display: inline-grid; grid-template-columns: auto auto; align-items: baseline; justify-content: center; }
.note__pitch > span { text-align: right; }
.note__pitch i {
  color: rgb(var(--v-theme-ink3));
  font-size: 0.42em;
  font-style: normal;
  font-weight: 400;
  text-align: left;
}

.note__hz {
  grid-area: hz;
  display: flex;
  justify-content: center;
  gap: 0.5ch;
  color: rgb(var(--v-theme-ink2));
  font-family: var(--font-mono);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  line-height: 18px;
}
.note__transpose { color: rgb(var(--v-theme-ink3)); }

.note__status {
  grid-area: status;
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.note--locked .note__status { color: rgb(var(--v-theme-ink)); font-weight: 700; }
.note__a4 { grid-area: a4; }

@media (max-width: 959px), (max-height: 559px) {
  .note {
    container-type: normal;
    grid-template-areas: "glyph status" "glyph hz" "glyph a4";
    grid-template-columns: 80px minmax(0, 1fr);
    grid-template-rows: 18px 18px 26px;
    align-content: center;
    column-gap: 8px;
    text-align: left;
  }
  .note__glyph { font-size: 32px; }
  .note__glyph--chord { font-size: 22px; }
  .note__hz { justify-content: start; }
  .note__a4 { justify-content: start; }
}
@media (max-height: 559px) and (min-width: 560px) {
  .note {
    grid-template-areas: "glyph status hz" "glyph a4 a4";
    grid-template-columns: 52px minmax(0, 1fr) 94px;
    grid-template-rows: 18px 26px;
    column-gap: 6px;
  }
  .note__glyph { font-size: 28px; }
  .note__glyph--chord { font-size: 17px; }
  .note__hz { font-size: 13px; justify-content: end; }
}
</style>
