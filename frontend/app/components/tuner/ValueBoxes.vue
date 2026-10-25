<template>
  <div class="boxes">
    <UiToolKey
      v-if="pages > 1"
      icon="mdi-chevron-left"
      :title="t('common.previous')"
      :disabled="page === 0"
      @click="page--"
    />
    <TunerValueBox
      v-for="(box, i) in shownBoxes"
      :key="`${box.kind}-${i}`"
      :box="box"
    />
    <UiToolKey
      v-if="pages > 1"
      icon="mdi-chevron-right"
      :title="t('common.next')"
      :disabled="page >= pages - 1"
      @click="page++"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { buildBoxes } from '~/utils/boxes'
import type { RankSlot } from '~/utils/boxes'
import { bassSlotsOf, octaveOf } from '~/utils/feet'
import { panelSlotIds, panelSlotOctaves } from '~/utils/panel'
import type { Bank } from '~/types/session'
import { useConfigSync } from '~/composables/useConfigSync'
import { useDeviationUnit } from '~/composables/useDeviationUnit'
import { useErrorReference } from '~/composables/useErrorReference'
import { useSessions } from '~/composables/useSessions'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const compact = useCompactLayout()
const page = ref(0)

// One box per slot plus a beat between each adjacent pair; flex row fits 1..8 slots (nothing assumes three).
// DOM not canvas, so the readout stays selectable and screen-readable.
const { panels, reedErrors, beatErrors, reeds, reedCount, mode, reedBoxes, octaveOffsets } = useTuner()
const { active } = useSessions()
const { config } = useConfigSync()
const unit = useDeviationUnit()
const { tolerance, beatTolerance } = useErrorReference()

// Box count follows the open session's bench.reeds; with no session, the engine's reedCount
// (live only while running) is the fallback.
const boxCount = computed(() =>
  active.value ? active.value.bench?.reeds ?? 0 : reedCount.value || 0,
)

// The row's rank slots, in the engine's slot order: the bass ladder's feet largest first when the
// bench faces the bass side, the pulled register's banks in card order otherwise.
const slots = computed<RankSlot[] | undefined>(() => {
  const bench = active.value?.bench
  if (!bench) return undefined
  if (bench.bass && bench.bassFeet?.length) {
    return bassSlotsOf(bench.bassFeet)
  }
  return (bench.banks as Bank[] | undefined)?.map(b => ({ name: b, octave: octaveOf(b) }))
})

// Without a session the panel decides; with one, the reed boxes the last analysis measured.
const ids = computed(() => {
  if (!active.value) return panelSlotIds(mode.value, reedBoxes.value, octaveOffsets.value)
  return reeds.value.length ? reeds.value.map(r => r.reed ?? 0) : undefined
})

const octaves = computed(() => {
  if (!active.value) return panelSlotOctaves(mode.value, reedBoxes.value, octaveOffsets.value)
  return slots.value?.map(s => s.octave) ?? reeds.value.map(r => r.octave ?? 0)
})

const boxes = computed(() => buildBoxes({
  panels: panels.value,
  reedErrors: reedErrors.value,
  beatErrors: beatErrors.value,

  // Passing the count keeps the row's shape even with nothing sounding.
  reedCount: boxCount.value,
  slots: slots.value,
  ids: ids.value,
  octaves: octaves.value,
  tones: !active.value && mode.value === 'chords',

  unit: unit.value,
  beatsInHz: config.tuner?.beat_unit !== 'reference',
  relative: config.tuner?.error_reference === 'goal',

  // tolerance only scales the bar; the in-tune verdict (inTol) is already decided by the backend.
  tolerance: tolerance.value,
  beatTolerance: beatTolerance.value,
}))
const pages = computed(() => compact.value ? Math.ceil(boxes.value.length / 3) : 1)
const shownBoxes = computed(() => compact.value ? boxes.value.slice(page.value * 3, page.value * 3 + 3) : boxes.value)
watch(pages, (count) => {
  page.value = Math.max(0, Math.min(page.value, count - 1))
})
</script>

<style scoped>
.boxes {
  display: flex;
  gap: var(--gap);
  min-height: 0;
}
.boxes > .key { align-self: center; flex: 0 0 auto; }
</style>
