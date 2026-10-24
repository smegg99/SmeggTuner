// Display conversion only: the engine, storage and goals carry reference units (1 = 0.06 percent of the note frequency).

export type DeviationUnit = 'reference' | 'cents'

/** core/dsp/meter.Missing: a slot or readout with no value. */
export const MISSING = -99

const UNIT_FRACTION = 0.0006

/** The tuner.deviation_unit setting, read; anything unknown is the reference unit. */
export function deviationUnit(setting: string | undefined): DeviationUnit {
  return setting === 'cents' ? 'cents' : 'reference'
}

/** core/target.Unit spelling, for the services that take a typed value with its unit. */
export function targetUnit(unit: DeviationUnit): 'reference' | 'cent' {
  return unit === 'cents' ? 'cent' : 'reference'
}

export function toCents(d: number): number {
  return 1200 * Math.log2(1 + UNIT_FRACTION * d)
}

/** A deviation from the scale, in the display unit. */
export function shownDeviation(d: number, unit: DeviationUnit): number {
  return unit === 'cents' ? toCents(d) : d
}

/** A residual (deviation minus base) in the display unit: in cents both ends convert, as core/target restates an error. */
export function shownResidual(residual: number, base: number, unit: DeviationUnit): number {
  return unit === 'cents' ? toCents(base + residual) - toCents(base) : residual
}

/** A curr/goal/error triple restated in the display unit, as core/target ReedError.In does. */
export function restate(e: { curr: number, goal: number, error: number }, unit: DeviationUnit) {
  if (unit === 'reference') return { curr: e.curr, goal: e.goal, error: e.error }
  const curr = toCents(e.curr)
  const goal = toCents(e.goal)
  return { curr, goal, error: curr - goal }
}
