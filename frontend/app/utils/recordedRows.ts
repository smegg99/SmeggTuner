import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import { restate } from '~/utils/deviation'
import type { DeviationUnit } from '~/utils/deviation'
import { beatColumns, beatGroups, byNote, reedGroups } from '~/utils/record'
import { rowKey } from '~/utils/recordOverview'

// Row models for the recorded table. Numbers are the backend's, in reference units; this restates them in the display unit and formats.

/** How a number is written, from useRecordFormat. Passed in so this stays pure. */
export interface RecordFormat {
  deviation: (v: number) => string
  signedDeviation: (v: number) => string
  /** a beat in hertz, unsigned, two decimals as the reference prints it */
  hz: (v: number) => string
  EMPTY: string
}

export interface RowOptions {
  /** beats in hertz (tuner.beat_unit), else in the deviation unit */
  beatsInHz?: boolean
  /** the take captured last (TableDTO.latest): the one Backspace removes */
  latest?: number
}

export interface ReedCell {
  key: string
  reed: number
  /** The measurement carries this reed; absent is not zero. */
  present: boolean
  /** Why an empty cell is empty: the take measured a slot in this column and it did not sound. */
  blank?: 'notHeard'
  curr: string
  goal: string
  error: string
  inTol: boolean
  /** The raw Curr, in the unit shown, so a cell can open its editor on it. */
  raw: number
  /** the take's slot this column reads, for RecordService.EditReed; -1 when none */
  slot: number
}

export interface BeatCell {
  key: string
  present: boolean
  curr: string
  goal: string
  error: string
  inTol: boolean
}

export interface RecordedRow {
  /** Keyed on the voice, not the note: two registers of one note are two rows. */
  key: string
  note: number
  noteName: string
  register?: string
  take: number
  manual: boolean
  /** this row holds the take captured last */
  latest: boolean
  reedCells: ReedCell[]
  beatCells: BeatCell[]
}

export function buildRecordedRows(
  rows: RowDTO[],
  reedCount: number,
  unit: DeviationUnit,
  fmt: RecordFormat,
  opts: RowOptions = {},
): RecordedRow[] {
  const reeds = reedGroups(reedCount)
  const beats = beatGroups(reedCount, rows)

  return byNote(rows).map((row) => {
    // A row that knows its columns places each slot there; without the mapping, by position.
    const colOf = (i: number) => row.cols?.[i] ?? i
    const measured = new Set(Array.from({ length: row.slots ?? 0 }, (_, slot) => colOf(slot)))

    const reedCells = reeds.map<ReedCell>((group) => {
      const reed = row.reeds.find(r => colOf(r.reed) === group.reed)
      const shown = reed ? restate(reed, unit) : undefined
      return {
        key: group.key,
        reed: group.reed,
        present: reed !== undefined,
        blank: reed === undefined && measured.has(group.reed) ? 'notHeard' : undefined,
        curr: shown ? fmt.signedDeviation(shown.curr) : fmt.EMPTY,
        goal: shown ? fmt.deviation(shown.goal) : fmt.EMPTY,
        error: shown ? fmt.signedDeviation(shown.error) : fmt.EMPTY,
        inTol: reed?.inTol ?? false,
        raw: shown?.curr ?? 0,
        slot: reed?.reed ?? -1,
      }
    })

    const beatCells = beats.map<BeatCell>((group) => {
      const beat = row.beats.find((b) => {
        const [low, high] = beatColumns(row, b)
        return low === group.low && high === group.high
      })
      const shown = beat ? restate(beat, unit) : undefined
      const curr = !beat ? fmt.EMPTY : opts.beatsInHz ? fmt.hz(beat.currHz) : fmt.signedDeviation(shown!.curr)
      return {
        key: group.key,
        present: beat !== undefined,
        curr,
        goal: shown ? fmt.deviation(shown.goal) : fmt.EMPTY,
        error: shown ? fmt.signedDeviation(shown.error) : fmt.EMPTY,
        inTol: beat?.inTol ?? false,
      }
    })

    return {
      key: rowKey(row),
      note: row.note,
      noteName: row.noteName,
      register: row.register,
      take: row.take,
      manual: row.manual,
      latest: row.take === opts.latest,
      reedCells,
      beatCells,
    }
  })
}
