<template>
  <div
    v-if="row"
    class="readings"
  >
    <UiPageStepper
      :model-value="index"
      :count="rows.length"
      :label="t('record.table.note')"
      @update:model-value="select"
    >
      <strong>{{ row.noteName }}</strong> {{ row.register }}
    </UiPageStepper>
    <UiPageStepper
      v-model="voice"
      :count="voiceCount"
      :label="voiceLabel"
    />
    <table class="reading-table">
      <thead><tr><th>{{ t('record.mobile.current') }}</th><th>{{ t('record.mobile.goal') }}</th><th>{{ t('record.mobile.error') }}</th></tr></thead>
      <tbody>
        <tr>
          <RecordRecordedReedCell
            v-if="reed"
            :cell="reed"
            :note-name="row.noteName"
            @edit="value => emit('edit', { take: row!.take, slot: reed!.slot, value })"
          />
          <td v-else>
            {{ beat?.curr }}
          </td>
          <td>{{ reed?.goal ?? beat?.goal }}</td><td>{{ reed?.error ?? beat?.error }}</td>
        </tr>
      </tbody>
    </table>
    <div class="readings__actions">
      <RecordRecordedRowMarks
        :row="row"
        :show-register="false"
      />
      <UiToolKey
        v-if="removable"
        icon="mdi-delete-outline"
        :title="t('record.table.delete')"
        @click="emit('delete-take', row.take)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { RecordedRow } from '~/utils/recordedRows'

const props = defineProps<{
  rows: RecordedRow[]
  reedHeads: { key: string, label: string }[]
  beatHeads: { key: string, label: string }[]
  selectedKey?: string | null
  removable: boolean
}>()
const emit = defineEmits<{
  'select': [string]
  'edit': [{ take: number, slot: number, value: number }]
  'delete-take': [number]
}>()
const { t } = useI18n()
const index = ref(0)
const voice = ref(0)
const row = computed(() => props.rows[index.value])
const voiceCount = computed(() => (row.value?.reedCells.length ?? 0) + (row.value?.beatCells.length ?? 0))
const reed = computed(() => row.value?.reedCells[voice.value])
const beat = computed(() => row.value?.beatCells[voice.value - (row.value?.reedCells.length ?? 0)])
const voiceLabel = computed(() => [...props.reedHeads, ...props.beatHeads][voice.value]?.label ?? '')
function select(next: number) {
  index.value = next
  const selected = props.rows[next]
  if (selected) emit('select', selected.key)
}
watch(() => props.selectedKey, (key) => {
  const next = props.rows.findIndex(r => r.key === key)
  if (next >= 0) index.value = next
}, { immediate: true })
watch(() => props.rows.length, (length) => {
  index.value = Math.max(0, Math.min(index.value, length - 1))
})
watch(voiceCount, (count) => {
  voice.value = Math.max(0, Math.min(voice.value, count - 1))
})
</script>

<style scoped>
.readings { display: flex; flex-direction: column; gap: 8px; min-height: 0; padding: 8px; }
.reading-table { table-layout: fixed; width: 100%; border-collapse: collapse; }
.reading-table th { font-size: 11px; font-weight: 500; color: rgb(var(--v-theme-ink3)); }
.reading-table td { padding: 12px 4px; text-align: center; font-variant-numeric: tabular-nums; }
.readings__actions { display: flex; justify-content: space-between; align-items: center; }
@media (max-height: 559px) {
  .readings { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 12px; }
  .reading-table { grid-column: 1; }
  .readings__actions { grid-column: 2; }
  .reading-table td { padding-block: 4px; }
}
@media (max-height: 359px) and (min-width: 560px) {
  .readings { grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr) 40px; grid-template-rows: 32px 32px; padding: 4px; gap: 4px; }
  .readings > .pages { grid-column: 1; }
  .readings > .pages :deep(.key) { width: 32px; min-width: 32px; height: 32px; min-height: 32px; }
  .reading-table { grid-column: 2; grid-row: 1 / 3; align-self: center; }
  .readings__actions { grid-column: 3; grid-row: 1 / 3; flex-direction: column; justify-content: center; }
}
</style>
