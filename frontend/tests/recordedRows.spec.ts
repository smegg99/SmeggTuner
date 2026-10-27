import { describe, expect, it } from 'vitest'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import { toCents } from '~/utils/deviation'
import { buildRecordedRows } from '~/utils/recordedRows'
import type { RecordFormat } from '~/utils/recordedRows'

/*
 * The recorded table's row models. A take's slots are positional (register bank order) and each
 * lands in its own card column; a slot that did not sound leaves its column empty and says why,
 * and two registers of one note stay two rows.
 */

const fmt: RecordFormat = {
  deviation: v => `${v}`,
  signedDeviation: v => (v > 0 ? `+${v}` : `${v}`),
  hz: v => v.toFixed(2),
  EMPTY: '-',
}

function reed(reed: number, curr: number, inTol: boolean): ReedError {
  return { reed, curr, goal: 0, error: curr, currHz: curr, goalHz: 0, errorHz: curr, inTol }
}

function beat(low: number, curr: number, inTol: boolean): BeatError {
  return { pair: `${low + 1}-${low + 2}`, low, high: low + 1, curr, goal: 0, error: curr, currHz: curr, goalHz: 0, errorHz: curr, inTol }
}

function row(note: number, opts: Partial<RowDTO> = {}): RowDTO {
  return {
    note, noteName: `n${note}`, take: 0, at: '', manual: false, slots: 2,
    reeds: [reed(0, -2, true), reed(1, 3, true)], beats: [beat(0, 4, false)],
    ...opts,
  }
}

describe('building the recorded rows', () => {
  /*
   * A row that knows its columns places each slot there: a bass take of a partial switch (16'
   * and 8' of a five-rank machine) must land its readings under those columns, not the first
   * two - and the beat between them travels with them.
   */
  it('places reeds by the row column mapping when it carries one', () => {
    const partial = row(45, { bass: true, cols: [1, 2] })
    const [r] = buildRecordedRows([partial], 5, 'reference', fmt)
    expect(r!.reedCells.map(c => c.present)).toEqual([false, true, true, false, false])
    expect(r!.reedCells[1]!.curr).toBe('-2')
    expect(r!.reedCells[2]!.curr).toBe('+3')
    expect(r!.beatCells.map(c => c.present)).toEqual([false, true, false, false])
    expect(r!.beatCells[1]!.curr).toBe('+4')
  })

  it('prints per-reed numbers for every slot that sounded', () => {
    const [r] = buildRecordedRows([row(60)], 2, 'reference', fmt)
    expect(r!.reedCells.map(c => c.curr)).toEqual(['-2', '+3'])
    expect(r!.reedCells.every(c => c.present && c.blank === undefined)).toBe(true)
  })

  // The take measured the slot and it did not sound: absent, which is not zero, and said so.
  it('leaves a silent slot its column, empty, and says why', () => {
    const silent = row(60, { reeds: [reed(0, -2, true)], beats: [] })
    const [r] = buildRecordedRows([silent], 2, 'reference', fmt)
    expect(r!.reedCells[1]!.present).toBe(false)
    expect(r!.reedCells[1]!.curr).toBe('-')
    expect(r!.reedCells[1]!.blank).toBe('notHeard')
    expect(r!.beatCells[0]!.present).toBe(false)
  })

  it('keeps one note on two registers as two distinct rows', () => {
    const rows = buildRecordedRows([row(60, { register: 'M2' }), row(60, { register: 'MMM' })], 2, 'reference', fmt)
    expect(rows).toHaveLength(2)
    expect(rows.map(r => r.register).sort()).toEqual(['M2', 'MMM'])
    // Keyed on the voice, not the note: keying on the note alone would collide them.
    expect(new Set(rows.map(r => r.key)).size).toBe(2)
  })

  // cents = 1200*log2(1 + 0.0006*d); an error restates as the difference of the two, as the report does.
  it('restates the numbers in cents when the setting says cents', () => {
    const r60 = row(60, { reeds: [{ ...reed(0, 10, true), goal: 5, error: 5 }], beats: [] })
    const [r] = buildRecordedRows([r60], 2, 'cents', fmt)
    expect(Number(r!.reedCells[0]!.curr)).toBeCloseTo(toCents(10))
    expect(Number(r!.reedCells[0]!.goal)).toBeCloseTo(toCents(5))
    expect(Number(r!.reedCells[0]!.error)).toBeCloseTo(toCents(10) - toCents(5))
    expect(r!.reedCells[0]!.raw).toBeCloseTo(toCents(10))
  })
})

/*
 * An LMH register on an LMMH card: three slots in columns L, M1 and H. The 8'-to-4' beat is
 * between M1 and H, and it has to print there, not in the musette column; the M2 column is not
 * one this register sounds, so it is empty without a claim.
 */
describe('a register short a rank', () => {
  const lmh = row(67, {
    register: 'LMH',
    slots: 3,
    cols: [0, 1, 3],
    reeds: [reed(0, 1.8, true), reed(1, -1.7, true), reed(2, 0.9, true)],
    beats: [beat(0, -3.4, true), beat(1, 2.6, true)],
  })

  it('prints the octave beat between the ranks it is between, not in the musette column', () => {
    const [built] = buildRecordedRows([lmh], 4, 'reference', fmt)
    const cells = Object.fromEntries(built!.beatCells.map(c => [c.key, c]))
    expect(cells['beat-1-2']?.present).toBe(false) // M1-M2, the musette column
    expect(cells['beat-1-3']?.present).toBe(true) // M1-H, where the reading belongs
    expect(cells['beat-1-3']?.curr).toBe('+2.6')
  })

  it('makes no claim about a column the register does not sound', () => {
    const [built] = buildRecordedRows([lmh], 4, 'reference', fmt)
    const m2 = built!.reedCells[2]!
    expect(m2.present).toBe(false)
    expect(m2.blank).toBeUndefined()
  })

  it('says not heard for a slot of the register that did not sound', () => {
    const quiet = { ...lmh, reeds: [reed(0, 1.8, true), reed(2, 0.9, true)], beats: [] }
    const [built] = buildRecordedRows([quiet], 4, 'reference', fmt)
    expect(built!.reedCells[1]!.blank).toBe('notHeard')
    expect(built!.reedCells[3]!.present).toBe(true)
  })
})

describe('the table\'s own switches', () => {
  // A beat box reads Hz by default; the table follows the same setting.
  it('prints beats in hertz when the beat unit is hertz', () => {
    const hz = row(60, { beats: [{ ...beat(0, 4, false), currHz: 1.234 }] })
    const [r] = buildRecordedRows([hz], 2, 'reference', fmt, { beatsInHz: true })
    expect(r!.beatCells[0]!.curr).toBe('1.23')
    const [u] = buildRecordedRows([hz], 2, 'reference', fmt)
    expect(u!.beatCells[0]!.curr).toBe('+4')
  })

  it('marks the row of the take captured last', () => {
    const rows = buildRecordedRows([row(60, { take: 0 }), row(62, { take: 1 })], 2, 'reference', fmt, { latest: 1 })
    expect(rows.map(r => r.latest)).toEqual([false, true])
  })

  // EditReed takes the take's slot, which a register short a rank places in another column.
  it('tells a cell which slot of its take it edits', () => {
    const lmh = row(67, { slots: 3, cols: [0, 1, 3], reeds: [reed(0, 1, true), reed(1, 2, true), reed(2, 3, true)], beats: [] })
    const [r] = buildRecordedRows([lmh], 4, 'reference', fmt)
    expect(r!.reedCells.map(c => c.slot)).toEqual([0, 1, -1, 2])
  })
})
