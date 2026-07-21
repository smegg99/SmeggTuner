<template>
  <span class="marks">
    <span
      v-if="showRegister && row.register"
      class="tag tag--mono"
    >{{ row.register }}</span>

    <span
      v-if="row.merged"
      class="mark"
    >
      <v-icon
        icon="mdi-sine-wave"
        size="x-small"
        class="mark--warn"
        :aria-label="t('record.table.merged')"
      />
      <v-tooltip
        activator="parent"
        location="top"
        max-width="360"
        :text="t('record.table.mergedHint')"
      />
    </span>
    <span
      v-else-if="row.derived"
      class="mark"
    >
      <v-icon
        icon="mdi-waveform"
        size="x-small"
        :aria-label="t('record.table.derived')"
      />
      <v-tooltip
        activator="parent"
        location="top"
        max-width="360"
        :text="t('record.table.derivedHint')"
      />
    </span>

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
  gap: 0.4cqw;
}

.tag {
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  flex: none;
  font-size: 1.35cqh;
  padding: 0 0.35cqw;
}

.tag--mono {
  font-family: var(--font-mono);
}

.mark {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: inline-flex;
}

.mark--warn {
  color: rgb(var(--v-theme-warn));
}
</style>
