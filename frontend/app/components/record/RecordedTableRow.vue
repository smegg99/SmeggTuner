<template>
  <tr
    class="recorded-row"
    :class="{ 'recorded-row--on': row.note === selectedNote }"
    @click="emit('select-note', row.note)"
  >
    <td class="note">
      <span class="note__name">{{ row.noteName }}</span>
      <RecordRecordedRowMarks
        :row="row"
        :show-register="showRegister"
      />
    </td>

    <template
      v-for="cell in row.reedCells"
      :key="cell.key"
    >
      <!-- Merged reeds show no value; the beat carries the reading. -->
      <td
        v-if="cell.merged"
        class="merged"
      >
        <v-icon
          icon="mdi-sine-wave"
          size="x-small"
        />
        {{ t('record.table.merged') }}
      </td>

      <template v-else>
        <td
          class="num num--curr"
          :class="{ 'num--out': cell.present && !cell.inTol }"
        >
          {{ cell.curr }}
        </td>
      </template>
    </template>

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
          <!-- Envelope-derived reading, not spectrum. -->
          <span
            v-if="beat.fromEnvelope"
            class="from-env"
          >
            <v-icon
              icon="mdi-waveform"
              size="x-small"
            />
            <v-tooltip
              activator="parent"
              location="top"
              max-width="360"
              :text="t('record.table.envelopeHint')"
            />
          </span>
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
import type { RecordedRow } from '~/utils/recordedRows'

defineProps<{
  row: RecordedRow
  removable: boolean
  showRegister: boolean
  selectedNote?: number | null
}>()

const emit = defineEmits<{
  'delete-take': [take: number]
  'select-note': [note: number]
}>()

const { t } = useI18n()
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

td {
  padding: 0.55cqh 0.7cqw;
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
  margin-inline-end: 0.5cqw;
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
  gap: 0.3cqw;
  justify-content: flex-end;
}

.from-env {
  color: rgb(var(--v-theme-ink3));
}

.merged {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.55cqh;
  text-align: center;
}

.kill {
  text-align: right;
  width: 3cqw;
}

.kill__key {
  background: none;
  border: 0;
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  opacity: 0;
  padding: 0.1cqh 0.2cqw;
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
