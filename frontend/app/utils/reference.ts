// The A4 and transposition controls: steps and the reasons services/tuner would refuse a change.

/** services/tuner: A4 430..450 Hz in 0.1 Hz steps; transposition 0..11 semitones. */
export const A4_MIN = 430
export const A4_MAX = 450
export const A4_STEP = 0.1
export const A4_DEFAULT = 440
export const TRANSPOSE_MAX = 11

export interface ReferenceGate {
  /** tuner.lock_a4 */
  lockA4: boolean
  /** a session is open; its reference and transposition are the ones being changed */
  session: boolean
  /** the open session's readings: measured against the reference, so it freezes */
  readings: number
  /** the panel mode, deciding only without a session */
  mode: string
}

/** a4Refusal is the i18n key of why A4 cannot move now, or '' when it can. */
export function a4Refusal(g: ReferenceGate): string {
  if (g.lockA4) return 'tuner.reference.a4Locked'
  if (g.session && g.readings > 0) return 'tuner.reference.frozen'
  return ''
}

/** transposeRefusal is the same for the transposition: chords and octaves run untransposed. */
export function transposeRefusal(g: ReferenceGate): string {
  if (g.session) return g.readings > 0 ? 'tuner.reference.frozen' : ''
  return g.mode === 'tones' ? '' : 'tuner.reference.transposeMode'
}

/** parseA4 reads a typed A4 (a comma counts as the decimal point) onto the 0.1 Hz grid, or null outside the range. */
export function parseA4(text: string): number | null {
  const hz = Number(text.trim().replace(',', '.'))
  if (!Number.isFinite(hz) || text.trim() === '') return null
  const rounded = Math.round(hz * 10) / 10
  return rounded >= A4_MIN && rounded <= A4_MAX ? rounded : null
}

export function stepTranspose(t: number, delta: number): number {
  return Math.min(TRANSPOSE_MAX, Math.max(0, t + delta))
}
