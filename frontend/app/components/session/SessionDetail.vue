<template>
  <div class="read">
    <SessionIdent
      v-model:view="view"
      :session="active"
      :src="src"
      :busy="busy"
    />

    <div
      v-if="view === 'recorded'"
      class="read__recorded"
    >
      <div
        v-if="!table"
        class="read__empty"
      >
        <p class="read__sub">
          {{ t('session.readings.empty') }}
        </p>
      </div>

      <!-- One table per recording: mixing recordings would mix separate takes of the accordion. -->
      <!-- Tag must be RecordRecordedTable, not RecordedTable: the short form resolves to a native element that silently renders nothing. -->
      <RecordRecordedTable
        v-else
        v-model:unit="unit"
        class="read__table"
        :rows="rows"
        :reed-count="table.reedCount"
        :banks="table.banks"
        :removable="true"
        :selected-note="selectedNote"
        @select-note="(n: number) => selectedNote = n"
        @delete-take="askDelete"
      />
    </div>

    <CurveEditor
      v-else
      class="read__pane"
    />

    <UiConfirmDialog
      v-model="confirmingDelete"
      :title="t('record.table.deleteTitle')"
      :body="t('record.table.deleteHint')"
      @confirm="confirmDelete"
    />

    <RecordReportDialog
      v-model="printing"
      :busy="reportBusy"
      :error="reportError ? t(reportError) : ''"
      @export="exportReport"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import SessionIdent from './SessionIdent.vue'
import { useRecord } from '~/composables/useRecord'
import { useReport } from '~/composables/useReport'
import { useSessions } from '~/composables/useSessions'
import { useWorkshop } from '~/composables/useWorkshop'
import type { ReportOptions, RecordUnit, TakeRow } from '~/types/record'

const props = defineProps<{
  id: string
  src: string
}>()

const { t } = useI18n()
const { active, busy } = useSessions()
const { table, deleteTake, show: showTable } = useRecord()
const { seq, last } = useWorkshop()
const {
  busy: reportBusy,
  error: reportError,
  exportSession,
  clearError,
} = useReport()

const unit = ref<RecordUnit>('cent')
const selectedNote = ref<number | null>(null)
const view = ref<'recorded' | 'curve'>('recorded')
const printing = ref(false)
const killTake = ref<number | null>(null)

const rows = computed<TakeRow[]>(() => (table.value?.rows ?? []) as unknown as TakeRow[])

// Refetch per id: covers first paint and clears the stale selection when switching sessions.
watch(() => props.id, () => {
  selectedNote.value = null
  void showTable()
}, { immediate: true })

// Responds to the toolbar Print action; disabled upstream when there's nothing to print.
watch(seq, () => {
  if (last.value !== 'print') return
  clearError()
  printing.value = true
})

async function exportReport(options: ReportOptions) {
  // Success closes the dialog; cancel or failure leaves it open with the error shown.
  if (await exportSession(options)) printing.value = false
}

function askDelete(take: number) {
  killTake.value = take
}

// Adapts the boolean dialog state to the killTake number.
const confirmingDelete = computed({
  get: () => killTake.value !== null,
  set: (open: boolean) => {
    if (!open) killTake.value = null
  },
})

async function confirmDelete() {
  const take = killTake.value
  killTake.value = null
  if (take === null) return
  // deleteTake refetches: removing a take reindexes the rest, so a stale table would target the wrong reading.
  await deleteTake(take)
}
</script>

<style scoped>
.read {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
}

/* Table must fill the pane, not sit content-height at the top. */
.read__recorded {
  display: grid;
  grid-template-rows: minmax(0, 1fr);
  min-height: 0;
}

.read__pane {
  min-height: 0;
  overflow: hidden;
}

/* Occupies the table's row when no recording is picked. */
.read__empty {
  align-items: center;
  display: flex;
  justify-content: center;
  text-align: center;
}

.read__sub {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  margin: 0;
  max-width: 46ch;
}

.read__table {
  min-height: 0;
  overflow: hidden;
}
</style>
