<template>
  <div class="panel">
    <RecordedTableToolbar
      :count="rowViews.length"
      @help="helpOpen = true"
    />

    <div
      v-if="!rowViews.length"
      class="empty"
    >
      {{ t('record.table.empty') }}
    </div>

    <div
      v-else
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
            :selected-note="selectedNote"
            @delete-take="emit('delete-take', $event)"
            @select-note="emit('select-note', $event)"
          />
        </tbody>
      </table>
    </div>

    <RecordedTableHelp v-model="helpOpen" />
  </div>
</template>

<script setup lang="ts">
// Row models are built in ~/utils/recordedRows; editing a Curr emits so the backend recomputes.
import { computed, ref } from 'vue'
import type { RecordUnit, TakeRow } from '~/types/record'
import { beatGroups, reedGroups } from '~/utils/record'
import { buildRecordedRows } from '~/utils/recordedRows'
import { useRecordFormat } from '~/composables/useRecordFormat'
import RecordedTableToolbar from './RecordedTableToolbar.vue'
import RecordedTableHelp from './RecordedTableHelp.vue'

const props = withDefaults(defineProps<{
  rows: TakeRow[]
  /** Reeds this register sounds, 1..8; never assume three. */
  reedCount: number
  /** Ranks in card order, naming the reed columns; numbered when absent. The treble side passes
   * its banks (L, M1...), the bass side its feet ("32'", "16'"...). */
  banks?: readonly string[]
  removable?: boolean
  selectedNote?: number | null
}>(), {
  banks: undefined,
  removable: true,
  selectedNote: null,
})

const emit = defineEmits<{
  'delete-take': [take: number]
  'select-note': [note: number]
}>()

const unit = defineModel<RecordUnit>('unit', { default: 'cent' })

const { t } = useI18n()

const helpOpen = ref(false)
const fmt = useRecordFormat()

const rowViews = computed(() => buildRecordedRows(props.rows, props.reedCount, unit.value === 'hz', fmt))

const reedHeads = computed(() => reedGroups(props.reedCount).map(g => ({ key: g.key, label: headOf(g.reed) })))
const beatHeads = computed(() => beatGroups(props.reedCount).map(g => ({ key: g.key, label: t('record.table.beat', { pair: g.pair }) })))

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
.panel {
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.empty {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 1.6cqh;
  justify-content: center;
  padding: 4cqh;
}

.scroll {
  min-height: 0;
  overflow: auto;
}

.table {
  border-collapse: collapse;
  font-size: 1.75cqh;
  width: 100%;
}

.table :deep(td),
.table :deep(th) {
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
}
</style>
