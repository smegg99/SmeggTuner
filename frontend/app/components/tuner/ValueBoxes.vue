<template>
  <div class="boxes">
    <TunerValueBox
      v-for="(box, i) in boxes"
      :key="`${box.kind}-${i}`"
      :box="box"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { buildBoxes } from '~/utils/boxes'
import type { RankSlot } from '~/utils/boxes'
import { bassSlotsOf, octaveOf } from '~/utils/feet'
import type { Bank } from '~/types/session'
import { useConfigSync } from '~/composables/useConfigSync'
import { useSessions } from '~/composables/useSessions'
import { useTuner } from '~/composables/useTuner'

// One box per reed plus a beat between each pair; flex row fits 1..8 reeds (nothing assumes three).
// DOM not canvas, so the readout stays selectable and screen-readable.
const { reedErrors, beatErrors, reeds, bands, reedsSeparated, reedsFromBeat, reedCount } = useTuner()
const { active } = useSessions()
const { config } = useConfigSync()

// Box count follows the open session's bench.reeds; with no session, the engine's reedCount
// (live only while running) is the fallback.
const boxCount = computed(() =>
  active.value ? active.value.bench?.reeds ?? 0 : reedCount.value || 0,
)

// The row's rank slots: the bass ladder's feet when the bench faces the bass side, the pulled
// register's banks otherwise. The bindings' Bank enum carries the same strings the app union names.
const slots = computed<RankSlot[] | undefined>(() => {
  const bench = active.value?.bench
  if (!bench) return undefined
  if (bench.bass && bench.bassFeet?.length) {
    return bassSlotsOf(bench.bassFeet)
  }
  return (bench.banks as Bank[] | undefined)?.map(b => ({ name: b, octave: octaveOf(b) }))
})

const boxes = computed(() => buildBoxes({
  separated: reedsSeparated.value,
  fromBeat: reedsFromBeat.value,
  reeds: reedErrors.value,
  beats: beatErrors.value,

  // Passing the count keeps the row's shape even with nothing sounding.
  reedCount: boxCount.value,

  // The pulled register's ranks and the engine's octave tags: a register spanning octaves maps
  // each box onto its rank, so a silent 16' leaves the 16' box empty instead of shifting the row.
  slots: slots.value,
  octaves: reeds.value.map(r => r.octave ?? 0),
  bands: bands.value,

  // tolerance only scales the bar; the in-tune verdict (inTol) is already decided by the backend.
  tolerance: config.tuner?.tolerance ?? 0,
  beatTolerance: config.tuner?.beat_tolerance ?? 0,
}))
</script>

<style scoped>
.boxes {
  display: flex;
  gap: 0.85cqh;
  min-height: 0;
}
</style>
