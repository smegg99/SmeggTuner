// Chart.js options for the curve plot; theme colours come in resolved (the caller reads them live per repaint).
import type { Anchor } from '~/types/curve'

export interface CurveOptionsDeps {
  bounds: { min: number, max: number }
  ink: string
  line: string
  anchors: Anchor[]
  onPick: (reed: number, note: number | undefined) => void
  onCommit: (note: number, reed: number, value: number) => void
}

export function curveChartOptions(d: CurveOptionsDeps) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false as const, // the curve changes on Set, not per frame
    interaction: { mode: 'nearest' as const, intersect: true },
    onClick: (_e: unknown, els: { datasetIndex: number, index: number }[]) => {
      const hit = els[0]
      if (hit) d.onPick(hit.datasetIndex, d.anchors[hit.index]?.note)
    },
    plugins: {
      // Vertical only: sliding an anchor sideways would clear this note and set another.
      dragData: {
        round: 1,
        showTooltip: true,
        onDragEnd: (_e: unknown, datasetIndex: number, index: number, v: number) => {
          const at = d.anchors[index]
          if (at) d.onCommit(at.note, datasetIndex, v) // one write, on release
        },
      },
      legend: {
        display: true,
        position: 'bottom' as const,
        labels: { color: `rgb(${d.ink})`, boxWidth: 12, boxHeight: 2 },
      },
    },
    scales: {
      x: {
        grid: { color: `rgba(${d.line}, 0.6)`, drawTicks: false },
        border: { color: `rgba(${d.line}, 0.6)` },
        ticks: { color: `rgb(${d.ink})`, maxRotation: 0, autoSkip: true, maxTicksLimit: 10 },
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
