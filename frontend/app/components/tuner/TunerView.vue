<template>
  <div class="tuner">
    <FilePanel
      v-if="file"
      class="tuner__input"
    />
    <TunerInputPanel
      v-else
      class="tuner__input"
    />

    <TunerNoteRuler class="tuner__notes" />
    <TunerSpectrumPanel class="tuner__spectrum" />
    <TunerNotePanel class="tuner__note" />
    <TunerDeviationRuler class="tuner__ruler" />

    <TunerValueBoxes class="tuner__boxes" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'

// File view keys off the source selection, not the transport DTO: a slow/failed DTO round
// trip would otherwise render nothing with no error.
const { current } = useAudioDevices()

const file = computed(() => current.value.kind === 'file')
</script>

<style scoped>
/* Rows are window fractions so nothing needs scrolling. The deviation-ruler row is sized to
   keep its top/bottom readings from clipping (they run off below ~100px of canvas); the boxes
   size to their content (auto). */
.tuner {
  display: grid;
  gap: 0.85cqh;
  grid-template-areas:
    "input    input"
    "notes    notes"
    "spectrum note"
    "ruler    ruler"
    "boxes    boxes";
  grid-template-columns: 1fr 24cqw;
  grid-template-rows: 1fr 1fr 1fr 19cqh auto;
  min-height: 0;
  min-width: 0;
}

.tuner__input { grid-area: input; }
.tuner__notes { grid-area: notes; }
.tuner__spectrum { grid-area: spectrum; }
.tuner__note { grid-area: note; }
.tuner__ruler { grid-area: ruler; }

.tuner__boxes { grid-area: boxes; }
</style>
