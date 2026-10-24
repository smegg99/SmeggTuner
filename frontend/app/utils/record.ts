import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'

// core/session: reeds sounding per note, 1..8; the engine resolves up to six, and a hardcoded 3 anywhere is a bug.
export const MIN_REEDS = 1
export const MAX_REEDS = 8

// EMPTY is what a cell holds when a reed did not sound; deliberately not zero, which reads as a perfectly tuned reed.
export const EMPTY = '-'

export interface ReedGroup {
  kind: 'reed'
  key: string
  /** 0-based, as ReedError.reed is. */
  reed: number
}

export interface BeatGroup {
  kind: 'beat'
  key: string
  /** "1-2", "2-3": how BeatError.pair spells it, 1-based. */
  pair: string
  low: number
  high: number
}

export type ColumnGroup = ReedGroup | BeatGroup

export function reedCountOf(count: number): number {
  if (!Number.isFinite(count)) return MIN_REEDS
  return Math.min(MAX_REEDS, Math.max(MIN_REEDS, Math.trunc(count)))
}

/** The reeds an instrument of this many reeds has, 0-based. */
export function reedIndexes(count: number): number[] {
  return Array.from({ length: reedCountOf(count) }, (_, index) => index)
}

// columnGroups lists a comparison's columns left to right (reed, its beat with the next reed, reed...). Adjacent pairs only, else an eight-reed instrument is twenty-eight beat columns wide.
export function columnGroups(count: number): ColumnGroup[] {
  const out: ColumnGroup[] = []
  for (const reed of reedIndexes(count)) {
    if (reed > 0) {
      const low = reed - 1
      out.push({
        kind: 'beat',
        key: `beat-${low}`,
        pair: `${low + 1}-${reed + 1}`,
        low,
        high: reed,
      })
    }
    out.push({ kind: 'reed', key: `reed-${reed}`, reed })
  }
  return out
}

// reedGroups is the recorded table's layout: every reed first, then every beat - not columnGroups' order.
export function reedGroups(count: number): ReedGroup[] {
  return reedIndexes(count).map(reed => ({
    kind: 'reed',
    key: `reed-${reed}`,
    reed,
  }))
}

/** The two card columns a beat belongs between: its two slots' own. */
export function beatColumns(row: RowDTO, beat: BeatError): [number, number] {
  const col = (i: number) => row.cols?.[i] ?? i
  return [col(beat.low), col(beat.high)]
}

/** The adjacent pairs of an instrument of this many reeds, plus a column of its own for a beat
 * whose two ranks are next to each other in the READING but not on the card - a register short a
 * rank still beats across the gap, and that reading has nowhere else to go. None at all below two. */
export function beatGroups(count: number, rows?: readonly RowDTO[]): BeatGroup[] {
  const out: BeatGroup[] = []
  const seen = new Set<string>()
  const add = (low: number, high: number) => {
    const key = `beat-${low}-${high}`
    if (low < 0 || high <= low || seen.has(key)) return
    seen.add(key)
    out.push({ kind: 'beat', key, pair: `${low + 1}-${high + 1}`, low, high })
  }
  for (const reed of reedIndexes(count)) {
    if (reed > 0) add(reed - 1, reed)
  }
  for (const row of rows ?? []) {
    for (const beat of row.beats) {
      if (beat.high !== beat.low + 1) continue
      const [low, high] = beatColumns(row, beat)
      add(low, high)
    }
  }
  out.sort((a, b) => a.low - b.low || a.high - b.high)
  return out
}

// recordKeys: light.lit and light.disabled are independent - lit means readings are landing, disabled means no session; never both.
export type RecordHint = 'recording' | 'warmup' | 'noSession'

export interface RecordKeys {
  light: { lit: boolean, disabled: boolean, hint: RecordHint }
  undo: { disabled: boolean }
}

export function recordKeys(state: {
  session: boolean
  armed: boolean
  readings: number
  busy: boolean
}): RecordKeys {
  return {
    light: {
      lit: state.session && state.armed,
      disabled: !state.session,
      hint: !state.session ? 'noSession' : state.armed ? 'recording' : 'warmup',
    },
    // Not gated on arm: undo discards what was captured, not what's being captured now; services/session.UndoTake refuses only ErrNoSession.
    undo: { disabled: !state.session || state.readings === 0 || state.busy },
  }
}

export function reedOf(row: RowDTO, reed: number): ReedError | undefined {
  return row.reeds.find(candidate => candidate.reed === reed)
}

// byNote sorts a pass into reading order (note, then register), matching the backend; it does not collapse a note played on two registers.
export function byNote(rows: RowDTO[]): RowDTO[] {
  return [...rows].sort(
    (a, b) =>
      a.note - b.note || (a.register ?? '').localeCompare(b.register ?? ''),
  )
}
