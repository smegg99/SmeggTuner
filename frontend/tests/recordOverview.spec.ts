import { describe, expect, it } from 'vitest'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import { nextKey, overviewColumns, overviewSpan } from '~/utils/recordOverview'

const row = (note: number, register: string, errors: number[]): RowDTO => ({
  note, noteName: `n${note}`, register, take: 0, at: '', manual: false, slots: errors.length,
  reeds: errors.map((e, reed) => ({ reed, curr: e, goal: 0, error: e, hasGoal: false, currHz: 0, goalHz: 0, errorHz: 0, inTol: Math.abs(e) < 2 })),
  beats: [],
}) as RowDTO

describe('the record graph', () => {
  const rows = [row(62, 'MM', [1, -3]), row(60, 'M', [0.5]), row(62, 'M', [4])]

  it('has one column per note, holding every register of it in table order', () => {
    const cols = overviewColumns(rows, 'reference')
    expect(cols.map(c => c.note)).toEqual([60, 62])
    expect(cols[1]!.keys).toEqual(['62:M', '62:MM'])
    expect(cols[1]!.marks.map(m => m.inTol)).toEqual([false, true, false])
  })

  // The reference: a click selects a note's take, the next click its next take.
  it('cycles through a note\'s rows on repeated clicks', () => {
    const col = overviewColumns(rows, 'reference')[1]!
    expect(nextKey(col, null)).toBe('62:M')
    expect(nextKey(col, '62:M')).toBe('62:MM')
    expect(nextKey(col, '62:MM')).toBe('62:M')
    expect(nextKey(col, '60:M')).toBe('62:M')
  })

  it('is at least twice the tolerance high, in steps of five', () => {
    const cols = overviewColumns(rows, 'reference')
    expect(overviewSpan(cols, 1)).toBe(5)
    expect(overviewSpan(cols, 4)).toBe(10)
  })
})
