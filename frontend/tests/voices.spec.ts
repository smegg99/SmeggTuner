import { describe, expect, it } from 'vitest'
import { byNote } from '~/utils/record'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'

/*
 * A ROW IS A VOICE, NOT A NOTE.
 *
 * Middle C played on MMM and middle C played on M2 are readings of DIFFERENT REEDS - three
 * of them and one of them. The backend already makes them two rows; the frontend's job is to
 * put them in reading order and never to collapse them back together, because whichever one
 * it dropped would be a reading the technician took and can no longer see.
 */

function row(note: number, register?: string): RowDTO {
  return {
    note,
    noteName: String(note),
    register,
    take: 0,
    at: '',
    slots: 0,
    reeds: [],
    beats: [],
    manual: false,
  }
}

describe('the order a card is read in', () => {
  it('sorts by note first, whatever was played on what', () => {
    const got = byNote([row(72, 'M2'), row(60, 'MMM'), row(64, 'M2')])
    expect(got.map(r => r.note)).toEqual([60, 64, 72])
  })

  it('keeps two registers of one note as two rows', () => {
    const got = byNote([row(60, 'MMM'), row(60, 'M2')])
    expect(got).toHaveLength(2)
    expect(got.map(r => r.register)).toEqual(['M2', 'MMM'])
  })

  // A session on an instrument nobody has described has no register on any row, and it is
  // exactly the one row per note it always was.
  it('leaves an undescribed instrument alone', () => {
    const got = byNote([row(64), row(60), row(72)])
    expect(got.map(r => r.note)).toEqual([60, 64, 72])
    expect(got.every(r => r.register === undefined)).toBe(true)
  })
})
