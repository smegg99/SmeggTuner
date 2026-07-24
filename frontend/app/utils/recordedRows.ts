import type { TakeRow } from '~/types/record'
import { beatGroups, beatOf, byNote, reedGroups, reedOf, reedsDerived, reedsUsable } from '~/utils/record'

// Row models for the recorded table. Numbers are the backend's, carried in both units; this picks a unit and formats. A merged note without a beat carries no per-reed numbers (reedsUsable).

/** How a number is written, from useRecordFormat. Passed in so this stays pure. */
export interface RecordFormat {
  cents: (v: number) => string
  signedCents: (v: number) => string
  hertz: (v: number) => string
  signedHertz: (v: number) => string
  EMPTY: string
}

export interface ReedCell {
  key: string
  reed: number
  /** The measurement carries this reed; absent is not zero. */
  present: boolean
  /** Merged and not recovered from the beat: no numbers may be printed. */
  merged: boolean
  curr: string
  goal: string
  error: string
  inTol: boolean
  /** The raw Curr, in the unit shown, so a cell can open its editor on it. */
  raw: number
}

export interface BeatCell {
  key: string
  present: boolean
  curr: string
  goal: string
  error: string
  inTol: boolean
  fromEnvelope: boolean
}

export interface RecordedRow {
  /** Keyed on the voice, not the note: two registers of one note are two rows. */
  key: string
  note: number
  noteName: string
  register?: string
  take: number
  manual: boolean
  merged: boolean
  derived: boolean
  reedCells: ReedCell[]
  beatCells: BeatCell[]
}

export function buildRecordedRows(
  rows: TakeRow[],
  reedCount: number,
  asHz: boolean,
  fmt: RecordFormat,
): RecordedRow[] {
  const reeds = reedGroups(reedCount)
  const beats = beatGroups(reedCount)

  return byNote(rows).map((row) => {
    const usable = reedsUsable(row)
    // A row that knows its columns places each reed there; without the mapping, by position.
    const colOf = (i: number) => row.cols?.[i] ?? i

    const reedCells = reeds.map<ReedCell>((group) => {
      const reed = row.cols
        ? row.reeds.find(r => colOf(r.reed) === group.reed)
        : reedOf(row, group.reed)
      return {
        key: group.key,
        reed: group.reed,
        present: usable && reed !== undefined,
        merged: !usable,
        curr: reed ? (asHz ? fmt.signedHertz(reed.currHz) : fmt.signedCents(reed.curr)) : fmt.EMPTY,
        goal: reed ? (asHz ? fmt.hertz(reed.goalHz) : fmt.cents(reed.goal)) : fmt.EMPTY,
        error: reed ? (asHz ? fmt.signedHertz(reed.errorHz) : fmt.signedCents(reed.error)) : fmt.EMPTY,
        inTol: reed?.inTol ?? false,
        raw: reed ? (asHz ? reed.currHz : reed.curr) : 0,
      }
    })

    const beatCells = beats.map<BeatCell>((group) => {
      const beat = row.cols
        ? row.beats.find(b => colOf(b.low) === group.low && colOf(b.high) === group.high)
        : beatOf(row, group.pair)
      return {
        key: group.key,
        present: beat !== undefined,
        curr: beat ? (asHz ? fmt.signedHertz(beat.currHz) : fmt.signedCents(beat.curr)) : fmt.EMPTY,
        goal: beat ? (asHz ? fmt.hertz(beat.goalHz) : fmt.cents(beat.goal)) : fmt.EMPTY,
        error: beat ? (asHz ? fmt.signedHertz(beat.errorHz) : fmt.signedCents(beat.error)) : fmt.EMPTY,
        inTol: beat?.inTol ?? false,
        fromEnvelope: beat?.fromEnvelope ?? false,
      }
    })

    return {
      key: `${row.note}:${row.register ?? ''}`,
      note: row.note,
      noteName: row.noteName,
      register: row.register,
      take: row.take,
      manual: row.manual,
      merged: !usable,
      derived: reedsDerived(row),
      reedCells,
      beatCells,
    }
  })
}
