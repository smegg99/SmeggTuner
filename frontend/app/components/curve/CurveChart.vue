<template>
  <div class="panel">
    <UiPanelHead
      :title="t('sessionPage.tabs.curve')"
      :help-title="t('curve.helpTitle')"
      @help="emit('help')"
    >
      <span class="panel__unit">{{ t(unit === 'hz' ? 'record.table.hertz' : 'record.table.cent') }}</span>
    </UiPanelHead>

    <!-- Capture the pointer on the way down so a drag off the canvas edge is still delivered. -->
    <div
      class="plot"
      @pointerdown="capture"
    >
      <p
        v-if="!anchors.length"
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
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from 'vuetify'
import {
  CategoryScale,
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
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { Anchor, CurveUnit } from '~/types/curve'
import { noteName } from '~/utils/tuning'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, dragData)

const props = defineProps<{
  anchors: Anchor[]
  rankItems: ToolItem[]
  reed: number
  unit: CurveUnit
}>()

const emit = defineEmits<{
  pick: [{ reed: number, note: number | undefined }]
  commit: [{ note: number, reed: number, value: number }]
  help: []
}>()

const { t } = useI18n()
const theme = useTheme()

const reedCount = computed(() => props.rankItems.length)
const rankLabel = (i: number) => String(props.rankItems[i]?.label ?? '')

// Pin (don't fit) the y axis: a fitted axis rescales mid-drag; bounds round outward, zero always in.
const bounds = computed(() => {
  const values = props.anchors.flatMap(a =>
    Array.from({ length: reedCount.value }, (_, i) => a.reeds[i] ?? 0),
  )
  if (!values.length) return { min: -5, max: 5 }

  const step = 5
  const lo = Math.min(0, ...values)
  const hi = Math.max(0, ...values)
  return {
    min: Math.floor((lo - 2) / step) * step,
    max: Math.ceil((hi + 2) / step) * step,
  }
})

function capture(event: PointerEvent) {
  const el = event.target
  if (el instanceof HTMLCanvasElement) el.setPointerCapture(event.pointerId)
}

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
    labels: props.anchors.map(a => noteName(a.note)),
    datasets: Array.from({ length: reedCount.value }, (_, i) => {
      // Edited rank is accent, the rest ink: red/green already mean out-of/in-tolerance here.
      const mine = i === props.reed
      const color = mine ? `rgb(${accent})` : `rgba(${ink}, 0.55)`
      return {
        label: rankLabel(i),
        data: props.anchors.map(a => a.reeds[i] ?? 0), // short of reedCount = reed never set = zero

        borderColor: color,
        backgroundColor: color,
        borderWidth: mine ? 2.5 : 1.5,
        pointRadius: mine ? 4 : 3,
        pointHoverRadius: 7,
        pointHitRadius: 14, // drawn small, grabbed large
        tension: 0,
      }
    }),
  }
})

const chartOptions = computed(() => {
  void theme.global.name.value
  return curveChartOptions({
    bounds: bounds.value,
    ink: triplet('--v-theme-ink3', '115,123,136'),
    line: triplet('--v-theme-wellLine', '220,224,232'),
    anchors: props.anchors,
    onPick: (reed, note) => emit('pick', { reed, note }),
    onCommit: (note, reed, value) => emit('commit', { note, reed, value }),
  })
})
</script>

<style scoped>
.panel {
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
  overflow: hidden;
}

.panel__unit {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.5cqh;
}

/* Chart.js sizes to its container, so the container is the one with a size. */
.plot {
  height: 100%;
  min-height: 0;
  padding: 0.8cqh 0.9cqw;
  position: relative;
}

.plot__empty {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 1.6cqh;
  inset: 0;
  justify-content: center;
  margin: 0;
  position: absolute;
  text-align: center;
}
</style>
