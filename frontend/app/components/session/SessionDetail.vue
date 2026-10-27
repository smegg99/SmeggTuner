<template>
  <div class="read">
    <SessionIdent
      :view="view"
      :session="active"
      :busy="busy"
      @update:view="selectView"
    />

    <SessionReadings
      v-if="view === 'recorded'"
      v-model:selected="selectedKey"
      @delete-take="askDelete"
    />

    <CurveEditor
      v-else-if="view === 'curve'"
      class="read__pane"
    />

    <!-- Tag must be SessionAccordionPanel: a bare AccordionPanel resolves to a native
         element and silently renders nothing. -->
    <SessionAccordionPanel
      v-else
      class="read__accordion"
    />

    <ConfirmDialog
      v-model="confirmingDelete"
      :title="t('record.table.deleteTitle')"
      :message="t('record.table.deleteHint')"
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
import SessionReadings from './SessionReadings.vue'
import ConfirmDialog from '~/components/ui/ConfirmDialog.vue'
import { useRecord } from '~/composables/useRecord'
import { useReport } from '~/composables/useReport'
import { useSessions } from '~/composables/useSessions'
import { useWorkshop } from '~/composables/useWorkshop'
import type { ReportOptions } from '~/types/record'
import { flushEdits } from '~/composables/pendingEdits'
import { registerBackStep } from '~/composables/useBackNavigation'

const props = defineProps<{
  id: string
}>()

const { t } = useI18n()
const { active, busy } = useSessions()
const { deleteTake, show: showTable } = useRecord()
const { seq, last } = useWorkshop()
const {
  busy: reportBusy,
  error: reportError,
  exportSession,
  clearError,
} = useReport()

const selectedKey = ref<string | null>(null)
const view = ref<'recorded' | 'curve' | 'accordion'>('recorded')
async function selectView(next: typeof view.value) {
  if (await flushEdits()) view.value = next
}
registerBackStep(async () => {
  if (view.value === 'recorded') return false
  await selectView('recorded')
  return true
})
const printing = ref(false)
const killTake = ref<number | null>(null)

// Refetch per id: covers first paint and clears the stale selection when switching sessions.
watch(() => props.id, () => {
  selectedKey.value = null
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
  grid-template-columns: minmax(0, 1fr);
  min-width: 0;
  min-height: 0;
}

.read__pane {
  min-height: 0;
  overflow: hidden;
}

/* The accordion panel scrolls itself, so it must NOT get read__pane's overflow clamp: both
   classes land on the same element from two scoped stylesheets at equal specificity, and
   whichever is injected last wins. It was the clamp, and the form below the fold - the
   switch-name field - could not be reached. */
.read__accordion {
  min-height: 0;
}
</style>
