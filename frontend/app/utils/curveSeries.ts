import type { Anchor } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { GoalRowDTO, RecordedDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'
import { goalShown } from '~/utils/curveUnits'
import type { EntryUnit } from '~/utils/curveUnits'

// What the curve graph plots, per rank: the goal line the engine reads (gaps where there is none),
// the anchors that can be dragged, and the pass's readings laid over them.

export interface CurvePoint {
  x: number
  y: number | null
}

export type SeriesKind = 'goal' | 'anchor' | 'recorded'

export interface Series {
  kind: SeriesKind
  /** the rank (curve column), 0-based */
  reed: number
  points: CurvePoint[]
}

export interface SeriesInput {
  anchors: readonly Anchor[]
  goals: readonly GoalRowDTO[]
  recorded: readonly RecordedDTO[]
  ranks: number
  unit: EntryUnit
  showRecorded: boolean
}

// goalLine breaks the line wherever a note has no goal or the covered range skips notes.
function goalLine(goals: readonly GoalRowDTO[], reed: number, unit: EntryUnit): CurvePoint[] {
  const out: CurvePoint[] = []
  let last = Number.NaN
  for (const g of goals) {
    if (g.note - last > 1) out.push({ x: g.note - 0.5, y: null })
    const d = g.reeds[reed]
    out.push({ x: g.note, y: d === null || d === undefined ? null : goalShown(d, g.note, unit) })
    last = g.note
  }
  return out
}

export function curveSeries(input: SeriesInput): Series[] {
  const out: Series[] = []
  for (let reed = 0; reed < input.ranks; reed++) {
    out.push({ kind: 'goal', reed, points: goalLine(input.goals, reed, input.unit) })
    out.push({
      kind: 'anchor',
      reed,
      points: input.anchors.map(a => ({ x: a.note, y: goalShown(a.reeds[reed] ?? 0, a.note, input.unit) })),
    })
    if (!input.showRecorded) continue
    out.push({
      kind: 'recorded',
      reed,
      points: input.recorded.flatMap(r => r.reeds
        .filter(m => m.column === reed)
        .map(m => ({ x: r.note, y: goalShown(m.dev, r.note, input.unit) }))),
    })
  }
  return out
}

/** The notes the graph spans: everything it plots, with a note of room at each end. */
export function curveRange(series: readonly Series[]): { min: number, max: number } | null {
  const xs = series.flatMap(s => s.points.filter(p => p.y !== null).map(p => p.x))
  if (!xs.length) return null
  return { min: Math.floor(Math.min(...xs)) - 1, max: Math.ceil(Math.max(...xs)) + 1 }
}

// noteTicks labels the axis at every C when three fit, else every half octave, minor third or note, whichever first gives three.
export function noteTicks(min: number, max: number): number[] {
  for (const step of [12, 6, 3, 1]) {
    const ticks: number[] = []
    for (let n = Math.ceil(min / step) * step; n <= max; n += step) ticks.push(n)
    if (ticks.length >= 3 || step === 1) return ticks
  }
  return []
}

/** Every plotted value, for fitting the scale. */
export function seriesValues(series: readonly Series[]): number[] {
  return series.flatMap(s => s.points.flatMap(p => (p.y === null ? [] : [p.y])))
}
