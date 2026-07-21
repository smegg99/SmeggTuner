import { liveReedsUsable } from '~/utils/tuning'

// Values and verdicts come from the DTO as-is; the only number produced here is `frac` (drawing geometry).

/** One reed's or one beat's reading, as core/target computed it. */
export interface Reading {
  curr: number
  goal: number
  error: number
  inTol: boolean
}

export interface BoxInput {
  /** did the spectrum split the reeds */
  separated: boolean
  /** or were they recovered from the beat between them */
  fromBeat: boolean
  reeds: Reading[]
  beats: Reading[]
  /** technician's windows, for the bar's scale only */
  tolerance: number
  beatTolerance: number
  /** reeds the engine looks for; drives box count so the row is stable */
  reedCount: number
}

/** Why a box has no number. Never "it is zero". */
export type Blank
  /** nothing is sounding: no reading yet */
  = | 'idle'
  /** the spectrum could not split the pair, so those figures are lobes, not reeds */
    | 'merged'

export interface Box {
  kind: 'reed' | 'beat'
  /** which reed or pair, zero based; the label is composed via i18n elsewhere */
  index: number
  unit: 'cent' | 'Hz'
  /** null when there is no honest number. NEVER a fallback. */
  value: number | null
  goal: number | null
  /** set exactly when value is null, and it says WHY. */
  blank: Blank | null
  /** the backend's verdict, never re-derived here */
  out: boolean
  /** 0..1, where the error bar ends. Drawing geometry. */
  frac: number | null
}

// Track is four tolerances wide, so cents and hertz read identically and the notches sit in the same place whatever the technician sets.
const TRACK_TOLERANCES = 4

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

// Goal is track centre, bar length the error. Zero tolerance would divide by zero and a NaN width vanishes, so centre it.
function fraction(error: number, tolerance: number): number {
  if (!(tolerance > 0)) return 0.5

  return clamp(0.5 + (error / (tolerance * TRACK_TOLERANCES)) * 0.5, 0, 1)
}

const blankReed = (i: number, why: Blank): Box => ({
  kind: 'reed',
  index: i,
  unit: 'cent',
  value: null,
  goal: null,
  blank: why,
  // Unknown is not out of tolerance; the engine declined to give a verdict.
  out: false,
  frac: null,
})

const blankBeat = (i: number): Box => ({
  kind: 'beat',
  index: i,
  unit: 'Hz',
  value: null,
  goal: null,
  blank: 'idle',
  out: false,
  frac: null,
})

// The merged-pair rule lives in liveReedsUsable() (and record.ts), not duplicated here, so dial and report can't disagree.

// buildBoxes: one box per configured reed always, so the row keeps its places as voices drop out; an unheard reed is empty and says which kind of empty.
export function buildBoxes(input: BoxInput): Box[] {
  // Count is declared, not measured: the split count flickers each hop and would resize the flex row; only with nothing declared does the reading decide.
  const count = input.reedCount > 0 ? input.reedCount : input.reeds.length
  if (count <= 0) return []

  // Only meaningful once something is sounding.
  const merged = input.reeds.length > 0
    && !liveReedsUsable(input.separated, input.fromBeat)

  const boxes: Box[] = []

  for (let i = 0; i < count; i++) {
    if (i > 0) {
      const beat = input.beats[i - 1]

      boxes.push(beat
        ? {
            kind: 'beat',
            index: i,
            unit: 'Hz',
            // The beat survives a merged pair: measured off the envelope, not the spectrum.
            value: beat.error,
            goal: beat.goal,
            blank: null,
            out: !beat.inTol,
            frac: fraction(beat.error, input.beatTolerance),
          }
        : blankBeat(i))
    }

    const reed = input.reeds[i]

    if (!reed) {
      boxes.push(blankReed(i, 'idle'))
      continue
    }

    if (merged) {
      boxes.push(blankReed(i, 'merged'))
      continue
    }

    boxes.push({
      kind: 'reed',
      index: i,
      unit: 'cent',
      value: reed.error,
      goal: reed.goal,
      blank: null,
      out: !reed.inTol,
      frac: fraction(reed.error, input.tolerance),
    })
  }

  return boxes
}
