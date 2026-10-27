<template>
  <UiPanel
    class="record-panel"
    :title="t('record.table.label')"
    flush
  >
    <template #meta>
      <slot name="side" />
      <v-chip v-if="rowViews.length && !compact">
        {{ t('record.table.notes', rowViews.length) }}
      </v-chip>
      <v-chip v-if="transpose > 0 && !compact">
        {{ t('tuner.reference.transposed', { n: transpose }) }}
      </v-chip>
    </template>
    <template #actions>
      <RecordedTableToolbar
        v-model:graph="graph"
        @help="helpOpen = true"
      />
    </template>

    <div
      v-if="!rowViews.length"
      class="empty"
    >
      {{ t('record.table.empty') }}
    </div>

    <RecordRecordedOverview
      v-if="graph && rowViews.length"
      :rows="rows"
      :unit="unit"
      :tolerance="tolerance"
      :selected-key="selectedKey"
      @select="key => emit('select', key)"
    />

    <div
      v-if="rowViews.length && !compact"
      class="scroll"
    >
      <table class="table">
        <RecordRecordedTableHead
          :reed-heads="reedHeads"
          :beat-heads="beatHeads"
        />
        <tbody>
          <RecordRecordedTableRow
            v-for="row in rowViews"
            :key="row.key"
            :row="row"
            :removable="removable"
            :show-register="showRegister"
            :selected-key="selectedKey"
            @delete-take="emit('delete-take', $event)"
            @select="emit('select', $event)"
            @edit="edit"
          />
        </tbody>
      </table>
    </div>

    <RecordRecordedMobile
      v-if="compact && rowViews.length && !graph"
      :rows="rowViews"
      :reed-heads="reedHeads"
      :beat-heads="beatHeads"
      :selected-key="selectedKey"
      :removable="removable"
      @select="emit('select', $event)"
      @delete-take="emit('delete-take', $event)"
      @edit="edit"
    />

    <RecordedTableHelp v-model="helpOpen" />
  </UiPanel>
</template>

<script setup lang="ts">
// Row models are built in ~/utils/recordedRows; editing a Curr emits so the backend recomputes.
import { computed, ref } from 'vue'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import { beatGroups, reedGroups } from '~/utils/record'
import { buildRecordedRows } from '~/utils/recordedRows'
import { shownDeviation, targetUnit } from '~/utils/deviation'
import { useConfigSync } from '~/composables/useConfigSync'
import { useDeviationUnit } from '~/composables/useDeviationUnit'
import { useRecord } from '~/composables/useRecord'
import { useRecordFormat } from '~/composables/useRecordFormat'
import RecordedTableToolbar from './RecordedTableToolbar.vue'
import RecordedTableHelp from './RecordedTableHelp.vue'

const props = withDefaults(defineProps<{
  rows: RowDTO[]
  /** Reeds this register sounds, 1..8; never assume three. */
  reedCount: number
  /** Ranks in card order, naming the reed columns; numbered when absent. The treble side passes
   * its banks (L, M1...), the bass side its feet ("32'", "16'"...). */
  banks?: readonly string[]
  removable?: boolean
  /** the selected row's key (recordedRows), or null */
  selectedKey?: string | null
}>(), {
  banks: undefined,
  removable: true,
  selectedKey: null,
})

const emit = defineEmits<{
  'delete-take': [take: number]
  'select': [key: string]
}>()

const { t } = useI18n()
const { config } = useConfigSync()
const { table, editReed } = useRecord()
const compact = usePagedLayout()

const helpOpen = ref(false)
const graph = ref(false)
const fmt = useRecordFormat()
const unit = useDeviationUnit()

const transpose = computed(() => table.value?.transpose ?? 0)
const tolerance = computed(() => shownDeviation(table.value?.tolerance ?? 0, unit.value))

const rowViews = computed(() => buildRecordedRows(props.rows, props.reedCount, unit.value, fmt, {
  beatsInHz: config.tuner?.beat_unit !== 'reference',
  latest: table.value?.latest,
}))

// A typed Curr: the backend stores it at the slot's own pitch and marks the take as edited.
function edit(e: { take: number, slot: number, value: number }) {
  void editReed(e.take, e.slot, e.value, targetUnit(unit.value))
}

const reedHeads = computed(() => reedGroups(props.reedCount).map(g => ({ key: g.key, label: headOf(g.reed) })))
const beatHeads = computed(() => beatGroups(props.reedCount, props.rows).map(g => ({ key: g.key, label: t('record.table.beat', { pair: g.pair }) })))

// Use the rank's name only when banks line up with reedCount; otherwise the reed number.
function headOf(reed: number): string {
  const banks = props.banks
  if (banks && banks.length === props.reedCount && reed < banks.length) return banks[reed]!
  return t('record.table.reed', { index: reed + 1 })
}

// Name the register on a row only when the pass caught more than one.
const showRegister = computed(() => {
  const first = props.rows[0]?.register
  return props.rows.some(r => r.register !== first)
})
</script>

<style scoped>
.record-panel {
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.empty {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 11px;
  justify-content: center;
  padding: var(--gap-wide);
}

.scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
}

.table {
  border-collapse: collapse;
  font-size: 12px;
  width: 100%;
}

.table :deep(td),
.table :deep(th) {
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
}
@media (max-width: 959px), (max-height: 559px) {
  .record-panel :deep(.panel__title) { display: none; }
  .record-panel :deep(.panel__head) { justify-content: space-between; }
}
</style>
