<template>
  <UiPanel
    class="curve-panel"
    :title="t('sessionPage.tabs.curve')"
    flush
  >
    <template #actions>
      <slot name="tools" />
      <UiToolKey
        icon="mdi-information-outline"
        class="panel-help"
        :title="t('curve.helpTitle')"
        @click="emit('help')"
      />
    </template>

    <div
      class="plot"
    >
      <p
        v-if="!range"
        class="plot__empty"
      >
        {{ t('curve.startHint') }}
      </p>
      <Line
        v-else
        :data="chartData"
        :options="chartOptions"
      />
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from 'vuetify'
import {
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import dragData from 'chartjs-plugin-dragdata'
import { curveChartOptions } from '~/utils/curveChart'
import { curveBounds } from '~/utils/curveUnits'
import type { EntryUnit } from '~/utils/curveUnits'
import { curveRange, curveSeries, seriesValues } from '~/utils/curveSeries'
import type { ToolItem } from '@smegg99/s99ab/types'
import type { Anchor } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { GoalRowDTO, RecordedDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'

ChartJS.register(LinearScale, PointElement, LineElement, Tooltip, Legend, dragData)

const props = defineProps<{
  anchors: Anchor[]
  /** the goal line as the engine reads it (CurveView.goals) */
  goals: GoalRowDTO[]
  /** the pass's readings (CurveView.recorded); drawn only with showRecorded */
  recorded: RecordedDTO[]
  showRecorded: boolean
  rankItems: ToolItem[]
  reed: number
  unit: EntryUnit
  /** the graph's half height in the unit; 0 fits it to what is drawn */
  scale: number
}>()

const emit = defineEmits<{
  pick: [{ reed: number | undefined, note: number }]
  commit: [{ note: number, reed: number, value: number }]
  help: []
}>()

const { t } = useI18n()
const theme = useTheme()

const series = computed(() => curveSeries({
  anchors: props.anchors,
  goals: props.goals,
  recorded: props.recorded,
  ranks: props.rankItems.length,
  unit: props.unit,
  showRecorded: props.showRecorded,
}))

const range = computed(() => curveRange(series.value))
const bounds = computed(() => curveBounds(seriesValues(series.value), props.scale, props.unit))

// Live Vuetify "r,g,b" theme triplets; fallback only on prerender (no document), never painted.
function triplet(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback
}

const chartData = computed(() => {
  void theme.global.name.value // recompute on theme switch
  const accent = triplet('--v-theme-accent', '37,99,235')
  const ink = triplet('--v-theme-ink2', '63,70,83')

  return {
    datasets: series.value.map((s) => {
      // Edited rank is accent, the rest ink: red/green already mean out-of/in-tolerance here.
      const mine = s.reed === props.reed
      const color = mine ? `rgb(${accent})` : `rgba(${ink}, 0.55)`
      const label = String(props.rankItems[s.reed]?.label ?? '')
      if (s.kind === 'goal') {
        return { label, data: s.points, borderColor: color, backgroundColor: color, borderWidth: mine ? 2 : 1.2, pointRadius: 0, pointHitRadius: 0, spanGaps: false, tension: 0, dragData: false }
      }
      if (s.kind === 'recorded') {
        return { label, data: s.points, showLine: false, borderColor: color, backgroundColor: 'transparent', borderWidth: 1, pointRadius: 2.5, pointHoverRadius: 4, pointHitRadius: 4, dragData: false }
      }
      return { label, data: s.points, showLine: false, borderColor: color, backgroundColor: color, pointRadius: mine ? 4 : 3, pointHoverRadius: 7, pointHitRadius: 14 }
    }),
  }
})

const chartOptions = computed(() => {
  void theme.global.name.value
  return curveChartOptions({
    bounds: bounds.value,
    range: range.value,
    ink: triplet('--v-theme-ink3', '115,123,136'),
    line: triplet('--v-theme-wellLine', '220,224,232'),
    series: series.value,
    round: props.unit === 'hz' ? 2 : 1,
    onPick: (reed, note) => emit('pick', { reed, note }),
    onCommit: (note, reed, value) => emit('commit', { note, reed, value }),
  })
})
</script>

<style scoped>
.curve-panel {
  min-height: 0;
  overflow: hidden;
}

/* Chart.js sizes to its container, so the container is the one with a size. */
.plot {
  height: 100%;
  min-height: 0;
  padding: var(--gap);
  position: relative;
}

.plot__empty {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 11px;
  inset: 0;
  justify-content: center;
  margin: 0;
  position: absolute;
  text-align: center;
}
</style>
