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
    <!-- A compact window stands A4 beside the note instead (TunerView). -->
    <TunerA4Stepper
      v-if="!compact"
      class="note__a4"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useReadoutStatus, type ReadoutStatus } from '~/composables/useReadoutStatus'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const compact = useCompactLayout()
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
  grid-template-rows: minmax(0, 1fr) 18px 16px var(--s99ab-control-h-dense);
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

/* What the three rows under it leave, as a font size: 100cqh is the card. The plain size is for a
   WebView older than container units. The name is a line of text across the whole card, centred as
   text: a box shrunk to the name keeps, in WebKit, the width it had at an earlier font size when the
   card's height changes the size, and the name then wraps or sits off centre. */
.note__glyph {
  grid-area: glyph;
  display: grid;
  align-items: center;
  text-align: center;
  white-space: nowrap;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--v-theme-ink));
  font-size: 32px;
  font-size: clamp(26px, calc(100cqh - 56px - var(--s99ab-control-h-dense)), 88px);
  font-weight: 500;
  line-height: 1;
}
.note__glyph--chord { font-size: 22px; font-size: clamp(20px, calc((100cqh - 56px - var(--s99ab-control-h-dense)) * 0.6), 44px); }
.note__pitch i {
  color: rgb(var(--v-theme-ink3));
  font-size: 0.42em;
  font-style: normal;
  font-weight: 400;
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

/* Where the tuner gives the note a strip's height instead of a card's (portrait, and a landscape too
   short for the card): the note at the left, its state over its pitch beside it. */
@media (max-width: 959px) and (min-height: 560px), (max-width: 559px), (max-height: 399px) {
  .note {
    container-type: normal;
    grid-template-areas: "glyph status" "glyph hz";
    grid-template-columns: minmax(64px, auto) minmax(0, 1fr);
    grid-template-rows: 18px 18px;
    align-content: center;
    column-gap: var(--gap-wide);
    row-gap: 0;
    padding: 0 var(--gap-wide);
    text-align: left;
  }
  /* A fixed size cannot outgrow the strip, and a lowercase note name keeps its descender. */
  .note__glyph { font-size: 36px; overflow: visible; }
  .note__glyph--chord { font-size: 22px; }
  .note__hz { justify-content: start; }
}

/* A phone on its side keeps the card, read at a glance from a stand: the state heads it, the note takes
   all that is left, as large as the card's height and width allow, and its pitch closes the card in a
   size that carries as far. */
@media (min-width: 560px) and (min-height: 400px) and (max-height: 559px) {
  .note {
    grid-template-areas: "status" "glyph" "hz";
    grid-template-rows: 18px minmax(0, 1fr) 24px;
    padding: var(--gap);
  }
  .note__status { font-size: 13px; line-height: 18px; text-align: left; }
  .note__hz { font-size: 16px; line-height: 24px; }
  .note__glyph {
    font-size: 32px;
    font-size: clamp(32px, min(calc(100cqh - 48px), 32cqw), 120px);
  }
  .note__glyph--chord {
    font-size: 22px;
    font-size: clamp(20px, min(calc((100cqh - 48px) * 0.6), 18cqw), 72px);
  }
}

/* A landscape phone wide enough for the note keys in the top bar (TunerView): the note is an indicator
   in a row of keys, at their height: the strip's layout, smaller. */
@media (min-width: 780px) and (max-height: 559px) {
  .note {
    container-type: normal;
    grid-template-areas: "glyph status" "glyph hz";
    grid-template-columns: minmax(40px, auto) minmax(0, 1fr);
    grid-template-rows: 16px 18px;
    align-content: center;
    column-gap: 12px;
    row-gap: 0;
    padding: 0 12px;
    text-align: left;
  }
  .note__glyph { font-size: 26px; overflow: visible; }
  .note__glyph--chord { font-size: 18px; }
  .note__hz { justify-content: start; font-size: 13px; line-height: 18px; }
  .note__status { font-size: 11px; line-height: 16px; text-align: left; }
}
</style>
