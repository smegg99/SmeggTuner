// Chart.js options for the curve plot; theme colours come in resolved (the caller reads them live per repaint).
import type { Point, TooltipItem } from 'chart.js'
import { noteTicks } from '~/utils/curveSeries'
import type { Series } from '~/utils/curveSeries'
import { noteName } from '~/utils/tuning'

export interface CurveOptionsDeps {
  bounds: { min: number, max: number }
  range: { min: number, max: number } | null
  ink: string
  line: string
  /** the datasets in chart order, so a hit can be told apart */
  series: readonly Series[]
  /** a click on a note column picks it; a click on a point also picks its rank */
  onPick: (reed: number | undefined, note: number) => void
  onCommit: (note: number, reed: number, value: number) => void
  /** decimals a drag rounds to: 2 in Hz, else 1 */
  round: number
}

interface Hit { datasetIndex: number, index: number }
interface ClickEvent { x: number | null }
interface ChartLike { scales: { x?: { getValueForPixel: (px: number) => number | undefined } } }

export function curveChartOptions(d: CurveOptionsDeps) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false as const, // the curve changes on Set, not per frame
    interaction: { mode: 'nearest' as const, intersect: true },
    onClick: (e: ClickEvent, els: Hit[], chart: ChartLike) => {
      const hit = els[0]
      const s = hit ? d.series[hit.datasetIndex] : undefined
      const x = e.x === null ? undefined : chart.scales.x?.getValueForPixel(e.x)
      const note = s?.points[hit!.index]?.x ?? (x === undefined ? undefined : Math.round(x))
      if (note !== undefined) d.onPick(s?.reed, note)
    },
    plugins: {
      // Vertical only, anchors only: sliding sideways would clear this note and set another.
      dragData: {
        round: d.round,
        showTooltip: true,
        onDragEnd: (_e: unknown, datasetIndex: number, index: number, v: number | Point | null) => {
          const s = d.series[datasetIndex]
          const at = s?.points[index]
          const value = typeof v === 'number' ? v : v?.y
          if (s?.kind === 'anchor' && at && value != null) d.onCommit(at.x, s.reed, value)
        },
      },
      legend: {
        display: true,
        position: 'bottom' as const,
        labels: {
          color: `rgb(${d.ink})`,
          boxWidth: 12,
          boxHeight: 2,
          // One entry per rank: its anchors carry the name.
          filter: (item: { datasetIndex?: number }) => d.series[item.datasetIndex ?? -1]?.kind === 'anchor',
        },
      },
      tooltip: {
        callbacks: {
          title: (items: TooltipItem<'line'>[]) => (items[0]?.parsed.x != null ? noteName(items[0].parsed.x) : ''),
        },
      },
    },
    scales: {
      x: {
        type: 'linear' as const,
        min: d.range?.min,
        max: d.range?.max,
        grid: { color: `rgba(${d.line}, 0.6)`, drawTicks: false },
        border: { color: `rgba(${d.line}, 0.6)` },
        // Ticks on note names, not on whatever round numbers the scale would pick.
        afterBuildTicks: (axis: { min: number, max: number, ticks: { value: number }[] }) => {
          axis.ticks = noteTicks(axis.min, axis.max).map(value => ({ value }))
        },
        ticks: {
          color: `rgb(${d.ink})`,
          maxRotation: 0,
          autoSkip: false,
          callback: (v: number | string) => noteName(Number(v)),
        },
      },
      y: {
        min: d.bounds.min, // fixed, see bounds in the component
        max: d.bounds.max,
        grid: { color: `rgba(${d.line}, 0.6)`, drawTicks: false },
        border: { color: `rgba(${d.line}, 0.6)` },
        ticks: { color: `rgb(${d.ink})`, maxTicksLimit: 6 },
      },
    },
  }
}
