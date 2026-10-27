<template>
  <tr
    ref="el"
    class="recorded-row"
    :class="{ 'recorded-row--on': selected, 'recorded-row--latest': row.latest }"
    @click="emit('select', row.key)"
  >
    <td
      class="note"
      :title="row.latest ? t('record.table.latest') : undefined"
    >
      <span class="note__name">{{ row.noteName }}</span>
      <RecordRecordedRowMarks
        :row="row"
        :show-register="showRegister"
      />
    </td>

    <!-- An empty cell says why when the take knows: a slot it measured that did not sound. -->
    <RecordRecordedReedCell
      v-for="cell in row.reedCells"
      :key="cell.key"
      :cell="cell"
      :note-name="row.noteName"
      @edit="value => emit('edit', { take: row.take, slot: cell.slot, value })"
    />

    <template
      v-for="beat in row.beatCells"
      :key="beat.key"
    >
      <td
        class="num"
        :class="{ 'num--out': beat.present && !beat.inTol }"
      >
        <span class="num__wrap">
          {{ beat.curr }}
        </span>
      </td>
    </template>

    <td class="kill">
      <button
        v-if="removable"
        type="button"
        class="kill__key"
        :title="t('record.table.delete')"
        :aria-label="t('record.table.delete')"
        @click.stop="emit('delete-take', row.take)"
      >
        <v-icon
          icon="mdi-close"
          size="x-small"
        />
      </button>
    </td>
  </tr>
</template>

<script setup lang="ts">
// Out-of-tolerance (amber) is the backend's per-cell verdict, never re-judged here.
import { computed, useTemplateRef, watch } from 'vue'
import type { RecordedRow } from '~/utils/recordedRows'

const props = defineProps<{
  row: RecordedRow
  removable: boolean
  showRegister: boolean
  selectedKey?: string | null
}>()

const emit = defineEmits<{
  'delete-take': [take: number]
  'select': [key: string]
  'edit': [{ take: number, slot: number, value: number }]
}>()

const { t } = useI18n()

const el = useTemplateRef<HTMLTableRowElement>('el')
const selected = computed(() => props.row.key === props.selectedKey)

// A row picked on the graph may be out of view.
watch(selected, (on) => {
  if (on) el.value?.scrollIntoView({ block: 'nearest' })
})
</script>

<style scoped>
.recorded-row {
  cursor: pointer;
}

.recorded-row:hover {
  background: rgb(var(--v-theme-row));
}

.recorded-row--on {
  background: rgb(var(--v-theme-row));
}

/* The take captured last, the one Backspace removes: a rule down the note cell. */
.recorded-row--latest .note {
  box-shadow: inset 2px 0 0 rgb(var(--v-theme-accent));
}

td {
  padding: var(--gap);
  white-space: nowrap;
}

.note {
  background: rgb(var(--v-theme-well));
  left: 0;
  position: sticky;
  z-index: 1;
}

.recorded-row:hover .note,
.recorded-row--on .note {
  background: rgb(var(--v-theme-row));
}

.note__name {
  font-weight: 600;
  margin-inline-end: var(--gap);
}

.num {
  color: rgb(var(--v-theme-ink2));
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  text-align: right;
}

.num--out {
  color: rgb(var(--v-theme-warn));
}

.num__wrap {
  align-items: center;
  display: inline-flex;
  gap: var(--gap-tight);
  justify-content: flex-end;
}

.kill {
  text-align: right;
  width: 29px;
}

.kill__key {
  background: none;
  border: 0;
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  opacity: 0;
  padding: 0.8px 2px;
}

.recorded-row:hover .kill__key,
.recorded-row--on .kill__key,
.kill__key:focus-visible {
  opacity: 1;
}

.kill__key:hover {
  color: rgb(var(--v-theme-warn));
}
</style>
