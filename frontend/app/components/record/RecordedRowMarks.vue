<template>
  <span class="marks">
    <span
      v-if="showRegister && row.register"
      class="tag tag--mono"
    >{{ row.register }}</span>

    <span
      v-if="row.manual"
      class="mark"
    >
      <v-icon
        icon="mdi-pencil-outline"
        size="x-small"
        :aria-label="t('record.table.manual')"
      />
      <v-tooltip
        activator="parent"
        location="top"
        max-width="320"
        :text="t('record.table.manualHint')"
      />
    </span>
  </span>
</template>

<script setup lang="ts">
import type { RecordedRow } from '~/utils/recordedRows'

defineProps<{
  row: RecordedRow
  showRegister: boolean
}>()

const { t } = useI18n()
</script>

<style scoped>
.marks {
  align-items: center;
  display: inline-flex;
  gap: var(--gap-tight);
}

.tag {
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  flex: none;
  font-size: 11px;
  padding: var(--gap-tight);
}

.tag--mono {
  font-family: var(--font-mono);
}

.mark {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: inline-flex;
}
</style>
