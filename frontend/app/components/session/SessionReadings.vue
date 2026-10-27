<!-- app/components/session/SessionReadings.vue -->
<template>
  <!-- The session's recorded readings: the treble table, the bass side's own below it. -->
  <div class="read__recorded">
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
    <template v-else>
      <RecordRecordedTable
        v-if="(!compact || side === 'treble') && (trebleRows.length || !bassRows.length)"
        class="read__table"
        :rows="trebleRows"
        :reed-count="table.reedCount"
        :banks="table.banks"
        :removable="true"
        :selected-key="selected"
        @select="(k: string) => selected = k"
        @delete-take="(take: number) => emit('deleteTake', take)"
      >
        <template
          v-if="compact && bassRows.length && trebleRows.length"
          #side
        >
          <v-select
            v-model="side"
            :items="sideItems"
            item-title="label"
            item-value="value"
            class="read__side"
            :aria-label="t('tuner.side.hint')"
            hide-details
          />
        </template>
      </RecordRecordedTable>

      <!-- The bass side: its own columns, the machine's ranks by foot. -->
      <template v-if="bassRows.length && (!compact || side === 'bass' || !trebleRows.length)">
        <h3
          v-if="!compact"
          class="read__bass-title"
        >
          {{ t('record.table.bassTitle') }}
        </h3>
        <RecordRecordedTable
          class="read__table"
          :rows="bassRows"
          :reed-count="table.bassReedCount || bassReeds"
          :banks="bassFeetLabels"
          :removable="true"
          :selected-key="selected"
          @select="(k: string) => selected = k"
          @delete-take="(take: number) => emit('deleteTake', take)"
        >
          <template
            v-if="compact && bassRows.length && trebleRows.length"
            #side
          >
            <v-select
              v-model="side"
              :items="sideItems"
              item-title="label"
              item-value="value"
              class="read__side"
              :aria-label="t('tuner.side.hint')"
              hide-details
            />
          </template>
        </RecordRecordedTable>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRecord } from '~/composables/useRecord'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'

const selected = defineModel<string | null>('selected', { required: true })
const emit = defineEmits<{ deleteTake: [number] }>()

const { t } = useI18n()
const compact = useCompactLayout()
const { table } = useRecord()
const side = ref('treble')
const sideItems = computed(() => ['treble', 'bass'].map(value => ({ value, label: t(`tuner.side.${value}`) })))

const rows = computed<RowDTO[]>(() => table.value?.rows ?? [])
const trebleRows = computed(() => rows.value.filter(r => !r.bass))
const bassRows = computed(() => rows.value.filter(r => r.bass))

// The bass section's columns; a machine nobody declared falls back to the widest row.
const bassReeds = computed(() =>
  Math.max(1, ...bassRows.value.map(r => r.reeds.length)),
)
const bassFeetLabels = computed(() =>
  (table.value?.bassFeet ?? []).map((f: number) => `${f}'`),
)
</script>

<style scoped>
/* Table must fill the pane, not sit content-height at the top. */
.read__recorded {
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.read__bass-title {
  color: rgb(var(--v-theme-ink));
  font-size: 13.5px;
  margin: var(--gap) 0 var(--gap-tight);
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
  font-size: 11px;
  margin: 0;
  max-width: 46ch;
}

.read__table {
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
}
.read__side { width: 88px; flex: 0 0 88px; }
</style>
