<!-- app/components/tuner/TunerView.vue -->
<template>
  <div
    class="tuner"
    :class="{ 'tuner--compact': compact, 'tuner--signal': signalOpen }"
    :style="{ '--tuner-boxes': boxCount }"
  >
    <!-- The readout and its keys. Only a wide landscape phone lays this group out as a row of its own; in
         every other layout it and the band below hold nothing together and their children are the tuner's. -->
    <div class="tuner__head">
      <!-- A compact window stands A4 beside the note, a field of its own; a wide one keeps it in the note's card. -->
      <div class="tuner__readout">
        <TunerNotePanel
          v-show="!signalOpen"
          class="tuner__note"
        />
        <TunerA4Stepper
          v-if="compact"
          v-show="!signalOpen"
          class="tuner__a4"
          panel
        />
      </div>
      <!-- Compact windows give the note keys and the mode and reed keys a row each, wide ones carry
           them in the top bar, as does a landscape phone that has the width; an open session's register
           decides the mode and reeds. -->
      <div
        v-if="compact && !barTools"
        v-show="!signalOpen"
        class="tuner__tools"
      >
        <ShellToolbarNote />
        <ShellToolbarReading />
      </div>
      <TunerModePanel
        v-if="compact && !sessionReeds"
        v-show="!signalOpen"
        class="tuner__controls"
      />
      <UiToolGroup
        v-if="compact"
        v-show="!signalOpen"
        class="tuner__tabs"
        :model-value="plot"
        :items="plots"
        :label="t('tuner.spectrum.title')"
        @update:model-value="plot = $event as Plot"
      />
    </div>
    <FilePanel
      v-if="file"
      v-show="signalOpen || !compact || plot === 'input'"
      class="tuner__input"
    />
    <TunerInputPanel
      v-else
      v-show="signalOpen || !compact || plot === 'input'"
      class="tuner__input"
    />

    <!-- The signal generator keeps only the input; the readings are hidden, not torn down, so their
         canvases come back as they were. -->
    <TunerNoteRuler
      v-show="!signalOpen && (!compact || plot === 'notes')"
      class="tuner__notes"
    />
    <TunerSpectrumPanel
      v-show="!signalOpen && (!compact || plot === 'spectrum')"
      class="tuner__spectrum"
    />

    <div class="tuner__band">
      <SignalPanel
        v-if="signalOpen"
        class="tuner__signal"
      />
      <TunerValueBoxes
        v-else
        class="tuner__boxes"
      />
      <TunerDeviationRuler
        v-show="!signalOpen"
        class="tuner__ruler"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useWideLandscapeLayout } from '~/composables/useCompactLayout'
import { useSessions } from '~/composables/useSessions'
import { useSignal } from '~/composables/useSignal'
import { panelSlotIds } from '~/utils/panel'

// File view keys off the source selection, not the transport DTO: a slow/failed DTO round
// trip would otherwise render nothing with no error.
const { current } = useAudioDevices()
const { sessionReeds, mode, reedBoxes, octaveOffsets } = useTuner()
const { active } = useSessions()
const { open: signalOpen } = useSignal()
const { t } = useI18n()
const compact = useCompactLayout()
const barTools = useWideLandscapeLayout()
type Plot = 'input' | 'notes' | 'spectrum'
const plot = ref<Plot>('spectrum')
const plots = computed(() => [
  { value: 'input', label: t('tuner.inputPanel.title') },
  { value: 'notes', label: t('tuner.notes.title') },
  { value: 'spectrum', label: t('tuner.spectrum.title') },
])

// A box for each reed and a beat between each pair: a landscape phone shares its last row between the
// boxes and the ruler by this count.
const boxCount = computed(() => {
  const slots = active.value
    ? active.value.bench?.reeds ?? 0
    : panelSlotIds(mode.value, reedBoxes.value, octaveOffsets.value).length
  return Math.max(1, 2 * slots - 1)
})

const file = computed(() => current.value.kind === 'file')
watch(file, (value) => {
  if (value) plot.value = 'input'
})
</script>

<style scoped>
/* Rows are window fractions so nothing needs scrolling; minmax(0, 1fr) so no panel's content can
   push its row taller. The ruler is a thin strip and the boxes size to their fixed content. */
.tuner {
  display: grid;
  gap: var(--gap);
  grid-template-areas:
    "input    input"
    "notes    notes"
    "spectrum note"
    "ruler    ruler"
    "boxes    boxes";
  grid-template-columns: minmax(0, 1fr) 156px;
  grid-template-rows: repeat(3, minmax(0, 1fr)) 72px 90px;
  min-height: 0;
  min-width: 0;
}

.tuner__head,
.tuner__band,
.tuner__readout { display: contents; }
.tuner--compact .tuner__readout {
  display: flex;
  gap: var(--gap);
  grid-area: note;
  min-height: 0;
  min-width: 0;
}
.tuner--compact .tuner__note { flex: 1 1 auto; min-width: 0; }
.tuner--compact .tuner__a4 { flex: 0 0 auto; }

.tuner__input { grid-area: input; }
.tuner__notes { grid-area: notes; }
.tuner__spectrum { grid-area: spectrum; }
.tuner__note { grid-area: note; }
.tuner__ruler { grid-area: ruler; }

.tuner__boxes { grid-area: boxes; }
.tuner__tools { grid-area: tools; display: flex; justify-content: space-between; gap: var(--s99ab-control-gap); min-width: 0; }
.tuner__controls { grid-area: controls; }
.tuner__tabs { grid-area: tabs; }
.tuner__signal { grid-area: signal; }

/* The signal generator: the input keeps its size and the generator takes every row below it. */
.tuner--signal {
  grid-template-areas:
    "input  input"
    "signal signal"
    "signal signal"
    "signal signal"
    "signal signal";
}

/* Portrait has no fixed rows to keep: the input over the generator. Two classes, so the compact
   layouts below never override it. */
.tuner.tuner--compact.tuner--signal {
  grid-template-areas: "input" "signal";
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: minmax(0, 1fr) minmax(min-content, 2fr);
}
.tuner.tuner--compact.tuner--signal .tuner__input { grid-area: input; }

/* A landscape phone has no height for both: the generator takes the room. */
@media (max-height: 559px) {
  .tuner.tuner--compact.tuner--signal { grid-template-areas: "signal"; grid-template-rows: minmax(0, 1fr); }
  .tuner.tuner--compact.tuner--signal .tuner__input { display: none; }
}

/* Portrait: one column. The readout is a strip over the plot, which gets the full width; the plot's
   tab names it, so only the input keeps its header, for the level and the recording's keys. */
.tuner--compact {
  grid-template-areas:
    "note"
    "tools"
    "controls"
    "tabs"
    "plot"
    "ruler"
    "boxes";
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: var(--app-readout-h) var(--s99ab-control-h) auto var(--s99ab-control-h) minmax(0, 1fr) var(--app-readout-h) auto;
}
.tuner--compact .tuner__input,
.tuner--compact .tuner__notes,
.tuner--compact .tuner__spectrum { grid-area: plot; }
.tuner--compact .tuner__notes :deep(.panel__head),
.tuner--compact .tuner__spectrum :deep(.panel__head) { display: none; }
.tuner__tabs :deep(.key) { flex: 1 1 0; }

/* A landscape phone: the readout and its keys down the left, the plot beside them, and under both a
   band the boxes share with the ruler, a part to each box and three to the ruler. The band is as tall
   as a box whatever is chosen above it: the ruler takes that height and adds none of its own, for its
   scale grows with its width and would otherwise set the row by how many boxes sit beside it. */
@media (max-height: 559px) and (min-width: 560px) {
  .tuner--compact {
    grid-template-areas:
      "note     tabs"
      "note     plot"
      "tools    plot"
      "controls plot"
      "band     band";
    grid-template-columns: minmax(300px, 2fr) minmax(0, 3fr);
    grid-template-rows: var(--s99ab-control-h) minmax(0, 1fr) repeat(2, var(--s99ab-control-h)) auto;
  }
  .tuner--compact .tuner__band {
    display: flex;
    gap: var(--gap);
    grid-area: band;
    min-width: 0;
  }
  .tuner--compact .tuner__boxes { flex: var(--tuner-boxes) 1 0; min-width: 0; }
  .tuner--compact .tuner__ruler { flex: 3 1 0; min-width: 240px; contain: size; }
  /* The generator's layouts place the tuner's own children. */
  .tuner.tuner--compact.tuner--signal .tuner__band { display: contents; }
}

/* A landscape phone wide enough for the note keys in the top bar (useWideLandscapeLayout): the note is an
   indicator in one row with the mode keys and the plot's tabs, and the plot takes the whole width
   between that row and the band. */
@media (min-width: 780px) and (max-height: 559px) {
  .tuner--compact {
    grid-template-areas: "head" "plot" "band";
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: var(--s99ab-control-h) minmax(0, 1fr) auto;
  }
  .tuner--compact .tuner__head {
    display: flex;
    gap: var(--gap);
    grid-area: head;
    min-width: 0;
  }
  .tuner--compact .tuner__readout { flex: 1 1 300px; min-width: 280px; }
  .tuner--compact .tuner__controls { flex: 0 1 auto; min-width: 0; }
  .tuner--compact .tuner__tabs { flex: 1 1 210px; min-width: 0; }
  .tuner.tuner--compact.tuner--signal .tuner__head { display: contents; }
}
/* The generator hides the readout whole. */
.tuner.tuner--compact.tuner--signal .tuner__readout { display: none; }
</style>
