import type { Panel } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import { MISSING, shownDeviation, shownResidual } from '~/utils/deviation'
import type { DeviationUnit } from '~/utils/deviation'

// Values come from the engine's screen state and verdicts from core/target; the only number produced here is `frac` (drawing geometry) and the unit restatement.

/** One rank of the register being tuned: its display name and the octave it sounds at. The treble
 * side names banks (L, M1, H), the bass side feet (32', 16'). */
export interface RankSlot {
  name: string
  octave: number
}

export interface BoxInput {
  /** The reference readouts, reed, pair, reed, ... (dsp.Measurement.Panels). */
  panels: readonly Panel[]
  /** The last measured slots against the goal: the verdict and the bar. */
  reedErrors: readonly ReedError[]
  beatErrors: readonly BeatError[]
  /** technician's windows in reference units, for the bar's scale only */
  tolerance: number
  beatTolerance: number
  /** slots the bench declares; drives box count so the row is stable */
  reedCount: number
  /** the pulled register's ranks in slot order; a register spanning octaves names its boxes by them */
  slots?: readonly RankSlot[]
  unit: DeviationUnit
  /** the reference reed box of each slot (dsp.ReedMeasure.Reed); labels number by it */
  ids?: readonly number[]
  /** each slot's semitones from the key, to tell an octave pair from a gap between boxes */
  octaves?: readonly number[]
  /** chords: the slots are the tones of a chord */
  tones?: boolean
  /** pair readouts are beats in Hz (tuner.beat_unit), not deviation units */
  beatsInHz: boolean
  /** readouts are deviation minus target (tuner.error_reference goal) */
  relative: boolean
}

/** Why a box has no number. Never "it is zero". */
export type Blank
  /** nothing is sounding: no reading yet */
  = | 'idle'
  /** the register sounds this slot but the engine did not hear it */
    | 'notHeard'
  /** an octave pair: the reference reads its beat only in Hz */
    | 'hzOnly'

export type BoxUnit = DeviationUnit | 'hz'

export interface Box {
  kind: 'reed' | 'beat'
  /** which slot or pair, zero based; the label is composed via i18n elsewhere */
  index: number
  unit: BoxUnit
  /** null when there is no honest number. NEVER a fallback. */
  value: number | null
  goal: number | null
  /** the readout's own precision and sign convention */
  decimals: number
  signed: boolean
  /** no note: the numbers are the last ones, dimmed as the reference dims them */
  dimmed: boolean
  /** set exactly when value is null, and it says WHY. */
  blank: Blank | null
  /** the backend's verdict, never re-derived here */
  out: boolean
  /** 0..1, where the error bar ends. Drawing geometry. */
  frac: number | null
  /** the rank this box stands for, when the pulled register names one */
  rank?: string
  /** the reed box (or chord tone) a reed box is labelled with, and the two a beat is between */
  number?: number
  pair?: [number, number]
  /** a chord tone */
  tone?: boolean
  /** the readout's own annotation: the tone's note name in chords */
  text?: string
}

// Track is four tolerances wide, so the notches sit in the same place whatever the technician sets.
const TRACK_TOLERANCES = 4

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

// Goal is track centre, bar length the error. Zero tolerance would divide by zero and a NaN width vanishes, so centre it.
function fraction(error: number, tolerance: number): number {
  if (!(tolerance > 0)) return 0.5
  return clamp(0.5 + (error / (tolerance * TRACK_TOLERANCES)) * 0.5, 0, 1)
}

// readout is the panel's number, or null where the reference prints "-".
function readout(p: Panel | undefined): number | null {
  return p && p.value > MISSING && p.shown !== '-' ? p.value : null
}

function targetOf(p: Panel | undefined): number | null {
  return p && p.target > MISSING ? p.target : null
}

function box(
  kind: Box['kind'], index: number, p: Panel | undefined, unit: BoxUnit, relative: boolean,
  verdict: { error: number, inTol: boolean } | undefined, tolerance: number, blank: Blank,
): Box {
  const raw = readout(p)
  const goal = targetOf(p)
  const dimmed = p?.dimmed ?? false
  const shell = { kind, index, unit, decimals: p?.decimals ?? 1, signed: p?.signed ?? true, dimmed, text: p?.text || undefined }
  if (raw === null) {
    // Unknown is not out of tolerance; the engine declined to give a verdict.
    return { ...shell, value: null, goal: null, blank, out: false, frac: null }
  }
  const deviation = unit !== 'hz'
  const value = !deviation ? raw : relative && goal !== null ? shownResidual(raw, goal, unit) : shownDeviation(raw, unit)
  return {
    ...shell,
    value,
    goal: goal === null ? null : deviation ? shownDeviation(goal, unit) : goal,
    blank: null,
    out: !dimmed && verdict !== undefined && !verdict.inTol,
    frac: verdict ? fraction(verdict.error, tolerance) : null,
  }
}

// buildBoxes: one box per configured slot always, so the row keeps its places as voices drop out;
// an unheard slot is empty and says which kind of empty. Slots are positional, in register order.
export function buildBoxes(input: BoxInput): Box[] {
  const named = input.slots?.some(s => s.octave !== 0) ? input.slots : undefined
  // Count is declared, not measured; only with nothing declared does the engine's readout row decide.
  const count = named?.length || input.reedCount || Math.ceil(input.panels.length / 2)
  if (count <= 0) return []

  const reedPanel = (i: number) => input.panels[2 * i]
  const heard = Array.from({ length: count }, (_, i) => reedPanel(i))
    .some(p => readout(p) !== null && !p!.dimmed)
  const beatUnit: BoxUnit = input.beatsInHz ? 'hz' : input.unit
  const id = (i: number) => input.ids?.[i] ?? i + 1
  const tone = input.tones || undefined

  const boxes: Box[] = []
  for (let i = 0; i < count; i++) {
    if (i > 0) {
      const p = input.panels[2 * i - 1]
      const verdict = input.beatErrors.find(b => b.low === i - 1 && b.high === i)
      const beat = box('beat', i, p, beatUnit, input.relative, verdict, input.beatTolerance, 'idle')
      const labelled: Box = { ...beat, pair: [id(i - 1), id(i)], rank: named ? `${named[i - 1]!.name}-${named[i]!.name}` : undefined }
      // An absent pair has no box: reeds 1 and 3, chord tones. An octave pair in units says why it is empty.
      if (!p?.absent) boxes.push(labelled)
      else if (octavePair(input, i)) boxes.push({ ...labelled, value: null, goal: null, blank: 'hzOnly', out: false, frac: null })
    }
    const verdict = input.reedErrors.find(r => r.reed === i)
    const reed = box('reed', i, reedPanel(i), input.unit, input.relative, verdict, input.tolerance, heard ? 'notHeard' : 'idle')
    boxes.push({ ...reed, number: id(i), tone, rank: named ? named[i]!.name : undefined })
  }
  return boxes
}

// octavePair: the reference Bass mode hides an octave beat while beats are in units (core/dsp pairPanels).
function octavePair(input: BoxInput, i: number): boolean {
  const lo = input.octaves?.[i - 1]
  const hi = input.octaves?.[i]
  return !input.beatsInHz && lo !== undefined && hi !== undefined && Math.abs(hi - lo) === 12
}
