<template>
  <div class="curve-editor">
    <CurveChart
      :anchors="anchors"
      :rank-items="rankItems"
      :reed="reed"
      :unit="unit"
      @pick="pick"
      @commit="commit"
      @help="helpOpen = true"
    />

    <div class="bar">
      <UiToolGroup
        v-if="rankItems.length > 1"
        :model-value="reed"
        :items="rankItems"
        :label="t('curve.rank')"
        @update:model-value="v => reed = v as number"
      />

      <InstrumentNoteStepper v-model="note" />

      <UiNumberInput
        v-model="value"
        class="bar__value"
        :suffix="t(unit === 'hz' ? 'record.table.hertz' : 'record.table.cent')"
      />

      <UiToolKey
        icon="mdi-check"
        :label="t('curve.set')"
        :disabled="!note || busy"
        :title="t('curve.setHint')"
        @click="set"
      />
      <UiToolKey
        icon="mdi-close"
        :label="t('curve.remove')"
        :disabled="!hasAnchorHere || busy"
        :title="t('curve.removeHint')"
        @click="remove"
      />

      <span class="bar__sp" />

      <UiToolKey
        :label="t('curve.interpolate')"
        :active="interpolate"
        :title="t('curve.interpolateHint')"
        @click="setInterpolate(!interpolate)"
      />
      <UiToolKey
        icon="mdi-delete-outline"
        :label="t('curve.clearAll')"
        :disabled="!anchors.length || busy"
        :title="t('curve.clearAllHint')"
        @click="confirmClear = true"
      />
    </div>

    <UiConfirmDialog
      v-model="confirmClear"
      :title="t('curve.clearAllTitle')"
      :body="t('curve.clearAllConfirm')"
      @confirm="clearAll"
    />

    <CurveHelp v-model="helpOpen" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import CurveChart from './CurveChart.vue'
import CurveHelp from './CurveHelp.vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useSessions } from '~/composables/useSessions'

// The chart reads the curve's own joined anchors; nothing here computes a goal (the backend does).
const { t } = useI18n()
const {
  active,
  busy,
  setAnchor,
  clearAnchor,
  setInterpolate,
} = useSessions()

const reed = ref(0)
const note = ref(0)
const value = ref(0)
const confirmClear = ref(false)
const helpOpen = ref(false)

const curve = computed(() => active.value?.curve ?? null)
const anchors = computed(() => [...(curve.value?.anchors ?? [])].sort((a, b) => a.note - b.note))
const unit = computed(() => curve.value?.unit ?? 'cent')

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

const hasAnchorHere = computed(() => anchors.value.some(a => a.note === note.value))

function pick(hit: { reed: number, note: number | undefined }) {
  reed.value = hit.reed
  if (hit.note !== undefined) note.value = hit.note
}

function commit(edit: { note: number, reed: number, value: number }) {
  void setAnchor(edit.note, edit.reed, edit.value, unit.value)
}

watch([anchors, () => active.value?.instrument], () => {
  if (note.value) return
  const first = anchors.value[0]?.note
  const lo = active.value?.instrument?.lo ?? 0
  const hi = active.value?.instrument?.hi ?? 0
  note.value = first ?? (lo && hi && hi > lo ? Math.round((lo + hi) / 2) : 60)
}, { immediate: true })

// Picking a note that already has an anchor loads its value, so Set corrects rather than overwrites.
watch([note, reed], () => {
  const found = anchors.value.find(a => a.note === note.value)
  if (found) value.value = found.reeds[reed.value] ?? 0
})

async function set() {
  if (!note.value) return
  await setAnchor(note.value, reed.value, value.value, unit.value)
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
  gap: 0.85cqh;
  grid-template-rows: minmax(0, 1fr) auto;
  min-height: 0;
}

/* Matches ShellAppToolbar; no-wrap so Set/Remove never jump to a second line at some widths. */
.bar {
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 0.45cqh;
  display: flex;
  gap: 0.5cqw;
  min-width: 0;
  overflow: hidden;
  padding: 0.55cqh 0.65cqw;
}

/* Match the keys' 3.9cqh height so the row isn't crooked. */
.bar :deep(.num) {
  height: 3.9cqh;
}

.bar :deep(.num__input) {
  height: 100%;
}

.bar__value {
  flex: 0 0 auto;
  width: 9cqw;
}

.bar__sp {
  flex: 1 1 auto;
}
</style>
