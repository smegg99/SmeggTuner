<template>
  <!-- One column per note; a click selects its row in the table, a second click its next register. -->
  <svg
    class="graph"
    :viewBox="`0 0 ${width} ${HEIGHT}`"
    preserveAspectRatio="none"
    role="img"
    :aria-label="t('record.table.graphLabel')"
  >
    <line
      class="graph__zero"
      x1="0"
      :x2="width"
      :y1="HEIGHT / 2"
      :y2="HEIGHT / 2"
    />
    <g
      v-for="(col, i) in columns"
      :key="col.note"
      class="graph__col"
      :class="{ 'graph__col--on': selectedKey && col.keys.includes(selectedKey) }"
      @click="emit('select', nextKey(col, selectedKey))"
    >
      <title>{{ col.noteName }}</title>
      <rect
        class="graph__hit"
        :x="i * COL"
        y="0"
        :width="COL"
        :height="HEIGHT"
      />
      <rect
        v-for="(m, k) in col.marks"
        :key="k"
        class="graph__mark"
        :class="{ 'graph__mark--out': !m.inTol }"
        :x="i * COL + COL * 0.2"
        :y="yOf(m.error) - 1"
        :width="COL * 0.6"
        height="2"
      />
    </g>
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import type { DeviationUnit } from '~/utils/deviation'
import { nextKey, overviewColumns, overviewSpan } from '~/utils/recordOverview'

const props = defineProps<{
  rows: RowDTO[]
  unit: DeviationUnit
  /** the reed tolerance in the display unit; the graph is at least twice it high each way */
  tolerance: number
  selectedKey?: string | null
}>()

const emit = defineEmits<{ select: [key: string] }>()

const { t } = useI18n()

// viewBox units; the svg stretches to the panel width, so a column is as wide as there is room.
const HEIGHT = 40
const COL = 10

const columns = computed(() => overviewColumns(props.rows, props.unit))
const width = computed(() => Math.max(1, columns.value.length) * COL)
const span = computed(() => overviewSpan(columns.value, props.tolerance))

const yOf = (error: number) => HEIGHT / 2 - (Math.max(-span.value, Math.min(span.value, error)) / span.value) * (HEIGHT / 2 - 2)
</script>

<style scoped>
.graph {
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: block;
  height: 49px;
  width: 100%;
}

.graph__zero {
  stroke: rgb(var(--v-theme-ink3));
  stroke-width: 1;
  vector-effect: non-scaling-stroke;
}

.graph__col {
  cursor: pointer;
}

.graph__hit {
  fill: transparent;
}

.graph__col:hover .graph__hit,
.graph__col--on .graph__hit {
  fill: rgb(var(--v-theme-row));
}

.graph__mark {
  fill: rgb(var(--v-theme-ink2));
}

.graph__mark--out {
  fill: rgb(var(--v-theme-warn));
}
</style>
