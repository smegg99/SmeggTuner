import type { DeviationUnit } from '~/utils/deviation'
import { toCents } from '~/utils/deviation'

// The unit a beat list is shown and typed in: the deviation unit, or hertz as the reference beat list allows.

export type EntryUnit = DeviationUnit | 'hz'

/** core/dsp/meter.TargetHertzScale: 0.06 as a binary32, the reference's widened unit. */
const HZ_SCALE = Math.fround(0.06)

/** core/target goalA4: a goal in Hz is read at the note's pitch for A4 = 440 Hz, whatever the session's. */
const GOAL_A4 = 440

const pitch440 = (note: number) => GOAL_A4 * 2 ** ((note - 69) / 12)

/** goalIn restates a goal of d reference units at note in unit (core/target Unit.FromGoal). */
export function goalIn(d: number, note: number, unit: EntryUnit): number {
  if (unit === 'hz') return (d * pitch440(note) * HZ_SCALE) / 100
  return unit === 'cents' ? toCents(d) : d
}

/** goalShown rounds a restated goal to what the editor shows and drags: Hz to 0.01, the rest to 0.1. */
export function goalShown(d: number, note: number, unit: EntryUnit): number {
  const v = goalIn(d, note, unit)
  const grain = unit === 'hz' ? 100 : 10
  return Math.round(v * grain) / grain
}

/** The unit services/session.SetAnchor is told a typed value is in. */
export function anchorUnit(unit: EntryUnit): 'reference' | 'cent' | 'hz' {
  if (unit === 'hz') return 'hz'
  return unit === 'cents' ? 'cent' : 'reference'
}

/** The fixed half heights the graph's scale offers, per unit; 0 is fitted to the data. */
export const CURVE_SCALES: Record<EntryUnit, readonly number[]> = {
  reference: [0, 5, 10, 20, 50],
  cents: [0, 5, 10, 20, 50],
  hz: [0, 0.5, 1, 2, 5],
}

/** curveBounds: a fixed scale is symmetric; fitted, it rounds outward from the values with zero always in. */
export function curveBounds(values: readonly number[], scale: number, unit: EntryUnit): { min: number, max: number } {
  if (scale > 0) return { min: -scale, max: scale }
  const step = unit === 'hz' ? 0.5 : 5
  const pad = unit === 'hz' ? 0.2 : 2
  if (!values.length) return { min: -step, max: step }
  const lo = Math.min(0, ...values)
  const hi = Math.max(0, ...values)
  return { min: Math.floor((lo - pad) / step) * step, max: Math.ceil((hi + pad) / step) * step }
}
