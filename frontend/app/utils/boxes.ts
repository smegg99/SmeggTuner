import { liveReedsUsable } from '~/utils/tuning'
import { octaveOf } from '~/utils/feet'
import type { Bank } from '~/types/session'

// Values and verdicts come from the DTO as-is; the only number produced here is `frac` (drawing geometry).

/** One reed's or one beat's reading, as core/target computed it. */
export interface Reading {
  curr: number
  goal: number
  error: number
  inTol: boolean
  /** 0-based measurement-reed indices of a beat's pair; absent on reeds and in older fixtures */
  low?: number
  high?: number
}

/** One octave band's accounting from the engine, for a register that spans octaves. */
export interface BandInfo {
  octave: number
  ranks: number
  found: number
  ghostOnly: boolean
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
  /** the pulled register's banks in card order; a register spanning octaves maps boxes onto them */
  banks?: readonly Bank[]
  /** each measurement reed's octave, aligned with `reeds`; the compound engine sets them */
  octaves?: readonly number[]
  /** the engine's per-band accounting, for saying WHY a rank's box is empty */
  bands?: readonly BandInfo[]
}

/** Why a box has no number. Never "it is zero". */
export type Blank
  /** nothing is sounding: no reading yet */
  = | 'idle'
  /** the spectrum could not split the pair, so those figures are lobes, not reeds */
    | 'merged'
  /** the register sounds this rank but the engine did not hear it */
    | 'notHeard'
  /** all the engine heard in this rank's octave is the lower rank's harmonic - a blocked rank shows exactly this */
    | 'harmonicOnly'

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
  /** the rank this box stands for, when the pulled register names one */
  bank?: Bank
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

const beatBox = (i: number, beat: Reading, tolerance: number): Box => ({
  kind: 'beat',
  index: i,
  unit: 'Hz',
  value: beat.error,
  goal: beat.goal,
  blank: null,
  out: !beat.inTol,
  frac: fraction(beat.error, tolerance),
})

const reedBox = (i: number, reed: Reading, tolerance: number): Box => ({
  kind: 'reed',
  index: i,
  unit: 'cent',
  value: reed.error,
  goal: reed.goal,
  blank: null,
  out: !reed.inTol,
  frac: fraction(reed.error, tolerance),
})

// beatFor picks the beat between two measurement reeds by its pair indices; readings without pair
// info (older fixtures) fall back to the positional slot.
function beatFor(beats: readonly Reading[], lo: number, hi: number): Reading | undefined {
  const named = beats.find(b => b.low === lo && b.high === hi)
  if (named) return named
  const positional = beats[lo]
  return positional && positional.low === undefined ? positional : undefined
}

// The merged-pair rule lives in liveReedsUsable() (and record.ts), not duplicated here, so dial and report can't disagree.

// buildBoxes: one box per configured reed always, so the row keeps its places as voices drop out; an unheard reed is empty and says which kind of empty.
// A register spanning octaves maps each box onto its rank instead, so a missing 16' leaves the 16' box empty rather than shifting every reading one place left.
export function buildBoxes(input: BoxInput): Box[] {
  if (input.banks?.some(b => octaveOf(b) !== 0)) {
    return buildBankBoxes(input, input.banks)
  }

  // Count is declared, not measured: the split count flickers each hop and would resize the flex row; only with nothing declared does the reading decide.
  const count = input.reedCount > 0 ? input.reedCount : input.reeds.length
  if (count <= 0) return []

  // Only meaningful once something is sounding.
  const merged = input.reeds.length > 0
    && !liveReedsUsable(input.separated, input.fromBeat)

  const boxes: Box[] = []

  for (let i = 0; i < count; i++) {
    if (i > 0) {
      const beat = beatFor(input.beats, i - 1, i)
      boxes.push(beat ? beatBox(i, beat, input.beatTolerance) : blankBeat(i))
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

    boxes.push(reedBox(i, reed, input.tolerance))
  }

  return boxes
}

// buildBankBoxes lays the row out by the register's ranks: each reed claims the first unclaimed
// bank in its own octave (the same rule the card's columns follow), so every box keeps naming the
// same rank as voices come and go. In compound mode a found reed is a band's own line, never a
// merged lobe, so the merged blank does not apply; an empty box says notHeard, or harmonicOnly
// when the engine heard only the rank below's partial there - the blocked-rank case.
function buildBankBoxes(input: BoxInput, banks: readonly Bank[]): Box[] {
  const octaves = input.octaves ?? []
  const claimed = input.reeds.map(() => false)
  const slots = banks.map((bank) => {
    const octave = octaveOf(bank)
    let reed = -1
    for (let i = 0; i < input.reeds.length; i++) {
      if (!claimed[i] && (octaves[i] ?? 0) === octave) {
        claimed[i] = true
        reed = i
        break
      }
    }
    return { bank, octave, reed }
  })

  const idle = input.reeds.length === 0
  const boxes: Box[] = []
  slots.forEach((slot, s) => {
    if (s > 0) {
      const prev = slots[s - 1]!
      const beat = prev.reed >= 0 && slot.reed >= 0
        ? beatFor(input.beats, prev.reed, slot.reed)
        : undefined
      boxes.push(beat ? beatBox(s, beat, input.beatTolerance) : blankBeat(s))
    }

    if (slot.reed >= 0) {
      boxes.push({ ...reedBox(s, input.reeds[slot.reed]!, input.tolerance), bank: slot.bank })
      return
    }
    boxes.push({ ...blankReed(s, idle ? 'idle' : bandBlank(slot.octave, input.bands)), bank: slot.bank })
  })
  return boxes
}

// bandBlank is WHY a rank's box is empty while others read: the engine's per-band accounting says.
function bandBlank(octave: number, bands?: readonly BandInfo[]): Blank {
  const band = bands?.find(b => b.octave === octave)
  return band?.ghostOnly ? 'harmonicOnly' : 'notHeard'
}
