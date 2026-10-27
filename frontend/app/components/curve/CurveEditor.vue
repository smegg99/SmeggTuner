<template>
  <div class="curve-editor">
    <CurveChart
      :anchors="anchors"
      :goals="view?.goals ?? []"
      :recorded="view?.recorded ?? []"
      :show-recorded="showRecorded"
      :rank-items="rankItems"
      :reed="reed"
      :unit="entry"
      :scale="scale"
      @pick="pick"
      @commit="commit"
      @help="helpOpen = true"
    >
      <template #tools>
        <CurveTools
          v-if="!compact"
          v-model:recorded="showRecorded"
          v-model:scale="scale"
          v-model:hz="hz"
          :unit="unit"
        />
      </template>
    </CurveChart>

    <CurveBar
      v-model:reed="reed"
      v-model:note="note"
      v-model:value="value"
      v-model:recorded="showRecorded"
      v-model:scale="scale"
      v-model:hz="hz"
      :unit="unit"
      :entry="entry"
      :rank-items="rankItems"
      :interpolate="interpolate"
      :can-remove="hasAnchorHere"
      :can-clear="anchors.length > 0"
      :busy="busy"
      @set="set"
      @remove="remove"
      @interpolate="setInterpolate"
      @clear="confirmClear = true"
    />

    <ConfirmDialog
      v-model="confirmClear"
      :title="t('curve.clearAllTitle')"
      :message="t('curve.clearAllConfirm')"
      @confirm="clearAll"
    />

    <CurveHelp v-model="helpOpen" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import CurveBar from './CurveBar.vue'
import CurveChart from './CurveChart.vue'
import CurveHelp from './CurveHelp.vue'
import ConfirmDialog from '~/components/ui/ConfirmDialog.vue'
import CurveTools from './CurveTools.vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useCurveView } from '~/composables/useCurveView'
import { useDeviationUnit } from '~/composables/useDeviationUnit'
import { useSessions } from '~/composables/useSessions'
import { anchorUnit, goalShown } from '~/utils/curveUnits'
import type { EntryUnit } from '~/utils/curveUnits'

// The chart reads the curve's own joined anchors; nothing here computes a goal (the backend does).
// Anchors are stored in reference units and shown, typed and dragged in the entry unit; the backend converts.
const { t } = useI18n()
const compact = useCompactLayout()
const {
  active,
  busy,
  setAnchor,
  clearAnchor,
  setInterpolate,
} = useSessions()
const view = useCurveView()

const reed = ref(0)
const note = ref(0)
const value = ref(0)
const confirmClear = ref(false)
const helpOpen = ref(false)
const showRecorded = ref(false)
const scale = ref(0)

const curve = computed(() => active.value?.curve ?? null)
const anchors = computed(() => [...(curve.value?.anchors ?? [])].sort((a, b) => a.note - b.note))
const unit = useDeviationUnit()

// Hz entry is the reference beat list's other unit: read at A4 = 440 Hz, so it never moves with the session's A4.
const hz = ref(false)
const entry = computed<EntryUnit>(() => (hz.value ? 'hz' : unit.value))

// A scale picked in one unit means nothing in the other, and the typed value restates.
watch(entry, () => {
  scale.value = 0
})

// Absent is TRUE: core/target defaults it on, and a pre-flag curve is not one with it off.
const interpolate = computed(() => curve.value?.interpolate ?? true)

const reedCount = computed(() => active.value?.instrument?.reedCount ?? curve.value?.reedCount ?? 1)
const banks = computed(() => active.value?.instrument?.banks ?? [])

function rankLabel(index: number): string {
  const list = banks.value
  if (list.length === reedCount.value && index < list.length) return String(list[index])
  return t('record.table.reed', { index: index + 1 })
}

const rankItems = computed<ToolItem[]>(() =>
  Array.from({ length: reedCount.value }, (_, i) => ({ value: i, label: rankLabel(i) })),
)

const selectedAnchor = computed(() => anchors.value.find(a => a.note === note.value))
const selectedValue = computed(() => selectedAnchor.value?.reeds[reed.value])
const hasAnchorHere = computed(() => Boolean(selectedAnchor.value))

// A click on the graph is a click on that note's row: it selects the note, and the rank of a point hit.
function pick(hit: { reed: number | undefined, note: number }) {
  if (hit.reed !== undefined) reed.value = hit.reed
  note.value = hit.note
}

function commit(edit: { note: number, reed: number, value: number }) {
  void setAnchor(edit.note, edit.reed, edit.value, anchorUnit(entry.value))
}

watch([anchors, () => active.value?.instrument], () => {
  if (note.value) return
  const first = anchors.value[0]?.note
  const lo = active.value?.instrument?.lo ?? 0
  const hi = active.value?.instrument?.hi ?? 0
  note.value = first ?? (lo && hi && hi > lo ? Math.round((lo + hi) / 2) : 60)
}, { immediate: true })

// Picking a note that already has an anchor loads its value, so Set corrects rather than overwrites.
function load() {
  const found = selectedAnchor.value
  if (found) value.value = goalShown(found.reeds[reed.value] ?? 0, found.note, entry.value)
}

// Follow persisted edits (including a drag) without replacing a draft on unrelated session events.
watch([note, reed, entry, selectedValue], load, { immediate: true })

async function set() {
  if (!note.value) return
  // The field hands back text once typed in; the binding takes a number.
  await setAnchor(note.value, reed.value, Number(value.value), anchorUnit(entry.value))
}

async function remove() {
  if (!note.value) return
  await clearAnchor(note.value)
}

async function clearAll() {
  confirmClear.value = false
  for (const a of [...anchors.value]) await clearAnchor(a.note)
}
</script>

<style scoped>
.curve-editor {
  display: grid;
  gap: var(--gap);
  grid-template-rows: minmax(0, 1fr) auto;
  min-height: 0;
}
</style>
