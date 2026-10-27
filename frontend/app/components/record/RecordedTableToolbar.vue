<template>
  <div class="table-tools">
    <UiToolKey
      class="tuner-small"
      :label="compact ? undefined : t('record.table.graph')"
      :icon="compact ? 'mdi-chart-bar' : undefined"
      :active="graph"
      :title="t('record.table.graphHint')"
      @click="graph = !graph"
    />
    <UiToolGroup
      :model-value="beatUnit"
      :items="BEAT_UNITS"
      :label="t('settings.beatUnitLabel')"
      class="tuner-small"
      @update:model-value="v => config.tuner.beat_unit = String(v)"
    />
    <UiToolKey
      icon="mdi-information-outline"
      class="panel-help"
      :title="t('record.table.helpTitle')"
      @click="emit('help')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useConfigSync } from '~/composables/useConfigSync'

const emit = defineEmits<{ help: [] }>()

const graph = defineModel<boolean>('graph', { default: false })

const { t } = useI18n()
const compact = useCompactLayout()
const { config } = useConfigSync()

// The same setting as Settings > Beat unit: the table and the beat boxes read one unit.
const beatUnit = computed(() => (config.tuner?.beat_unit === 'reference' ? 'reference' : 'hz'))

const BEAT_UNITS = computed<ToolItem[]>(() => ['hz', 'reference'].map(value => ({ value, label: t(`tuner.unit.${value}`) })))
</script>

<style scoped>
/* The original panel head's control gap. */
.table-tools {
  align-items: center;
  display: flex;
  gap: var(--gap);
}
</style>
