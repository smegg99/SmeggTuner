<!-- app/components/tuner/TunerView.vue -->
<template>
  <div
    class="tuner"
    :class="{ 'tuner--compact': compact }"
  >
    <!-- Compact windows give the mode and reed keys a row, wide ones carry them in the top bar; an
         open session's register decides both. -->
    <TunerModePanel
      v-if="compact && !sessionReeds"
      class="tuner__controls"
    />
    <UiToolGroup
      v-if="compact"
      class="tuner__tabs"
      :model-value="plot"
      :items="plots"
      :label="t('tuner.spectrum.title')"
      @update:model-value="plot = $event as Plot"
    />
    <FilePanel
      v-if="file"
      v-show="!compact || plot === 'input'"
      class="tuner__input"
    />
    <TunerInputPanel
      v-else
      v-show="!compact || plot === 'input'"
      class="tuner__input"
    />

    <TunerNoteRuler
      v-show="!compact || plot === 'notes'"
      class="tuner__notes"
    />
    <TunerSpectrumPanel
      v-show="!compact || plot === 'spectrum'"
      class="tuner__spectrum"
    />
    <TunerNotePanel class="tuner__note" />
    <TunerDeviationRuler class="tuner__ruler" />

    <TunerValueBoxes class="tuner__boxes" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'

// File view keys off the source selection, not the transport DTO: a slow/failed DTO round
// trip would otherwise render nothing with no error.
const { current } = useAudioDevices()
const { sessionReeds } = useTuner()
const { t } = useI18n()
const compact = useCompactLayout()
type Plot = 'input' | 'notes' | 'spectrum'
const plot = ref<Plot>('spectrum')
const plots = computed(() => [
  { value: 'input', label: t('tuner.inputPanel.title') },
  { value: 'notes', label: t('tuner.notes.title') },
  { value: 'spectrum', label: t('tuner.spectrum.title') },
])

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

.tuner__input { grid-area: input; }
.tuner__notes { grid-area: notes; }
.tuner__spectrum { grid-area: spectrum; }
.tuner__note { grid-area: note; }
.tuner__ruler { grid-area: ruler; }

.tuner__boxes { grid-area: boxes; }
.tuner__controls { grid-area: controls; }
.tuner__tabs { grid-area: tabs; }

/* Portrait: one column. The readout is a short strip over the spectrum, which gets the full width. */
.tuner--compact {
    grid-template-areas:
      "note"
      "controls"
      "tabs"
      "plot"
      "ruler"
      "boxes";
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: 80px auto 40px minmax(0, 1fr) 64px 84px;
}
.tuner--compact .tuner__input,
.tuner--compact .tuner__notes,
.tuner--compact .tuner__spectrum { grid-area: plot; }
.tuner__tabs :deep(.key) { flex: 1 1 0; }

@media (max-height: 559px) and (min-width: 560px) {
  .tuner--compact {
    grid-template-areas: "note tabs" "controls plot" "boxes plot" "boxes ruler";
    grid-template-columns: minmax(260px, 2fr) minmax(0, 3fr);
    grid-template-rows: 48px 40px minmax(0, 1fr) 64px;
    align-items: stretch;
  }
  .tuner__controls { align-self: center; }
}
@media (max-width: 559px) and (max-height: 619px) {
  .tuner--compact { gap: 4px; grid-template-rows: 64px auto 40px minmax(0, 1fr) 52px 72px; }
}
@media (max-height: 359px) and (min-width: 560px) {
  .tuner--compact {
    grid-template-areas: "note tabs" "controls plot" "boxes plot" "boxes ruler";
    grid-template-rows: 44px 40px minmax(0, 1fr) 48px;
    gap: 4px;
  }
  .tuner__spectrum :deep(.panel__head), .tuner__notes :deep(.panel__head) { display: none; }
  .tuner__input:not(.file) :deep(.panel__head) { min-height: 0; padding-block: 0; }
  .tuner__input:not(.file) :deep(.panel__title) { display: none; }
}
</style>
