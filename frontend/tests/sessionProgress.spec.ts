import { describe, expect, it } from 'vitest'
import { sessionProgress } from '~/utils/sessionProgress'
import type { ProgressSession } from '~/utils/sessionProgress'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'

/*
 * The footer's session label, counted. The in-tune verdict is the backend's
 * (inTol); this asserts the label only sums it.
 */

function reed(inTol: boolean): ReedError {
  return { reed: 0, curr: 0, goal: 0, error: 0, currHz: 0, goalHz: 0, errorHz: 0, inTol }
}

function beat(inTol: boolean): BeatError {
  return { pair: '1-2', low: 0, high: 1, curr: 0, goal: 0, error: 0, currHz: 0, goalHz: 0, errorHz: 0, inTol }
}

function row(note: number, opts: Partial<RowDTO> = {}): RowDTO {
  return {
    note, noteName: '', take: 0, at: '', slots: 1,
    reeds: [reed(true)], beats: [], manual: false, ...opts,
  }
}

// 41-key keyboard (F3..A6-ish), a goal curve set: the full case.
function sess(opts: Partial<ProgressSession> = {}): ProgressSession {
  return { name: 'My Rossetti', instrument: { lo: 53, hi: 93 }, curve: { anchors: [{}] }, ...opts }
}

describe('the footer session label', () => {
  it('says nothing but "none" with no session', () => {
    expect(sessionProgress(null, [row(60)], true)).toEqual({ kind: 'none' })
  })

  it('shows only the name when no pass is being recorded', () => {
    expect(sessionProgress(sess(), [row(60)], false)).toEqual({ kind: 'idle', name: 'My Rossetti' })
  })

  it('counts caught notes against the keyboard, and how many are in tune', () => {
    // 23 notes caught, the first 18 inside tolerance.
    const rows = Array.from({ length: 23 }, (_, i) => row(53 + i, { reeds: [reed(i < 18)] }))
    expect(sessionProgress(sess(), rows, true)).toEqual({
      kind: 'active', name: 'My Rossetti', done: 23, total: 41, inTune: 18,
    })
  })

  it('drops the keyboard total when the instrument never said its span', () => {
    const rows = [row(60), row(61)]
    expect(sessionProgress(sess({ instrument: { lo: 0, hi: 0 } }), rows, true).total).toBeNull()
    expect(sessionProgress(sess({ instrument: null }), rows, true).total).toBeNull()
  })

  it('drops the in-tune count with no goal curve', () => {
    const out = sessionProgress(sess({ curve: { anchors: [] } }), [row(60)], true)
    expect(out).toEqual({ kind: 'active', name: 'My Rossetti', done: 1, total: 41, inTune: null })
  })

  it('shows no in-tune ratio before a note is caught', () => {
    expect(sessionProgress(sess(), [], true)).toEqual({
      kind: 'active', name: 'My Rossetti', done: 0, total: 41, inTune: null,
    })
  })

  // A row is judged by its reeds; a row with no slot heard is not a vacuous pass, whatever its beats say.
  it('never passes a row with no reed heard', () => {
    const outOfTune = row(60, { reeds: [reed(false)], beats: [beat(true)] })
    const noReading = row(61, { reeds: [], beats: [beat(true)] })
    const inTune = row(62)
    const out = sessionProgress(sess(), [outOfTune, noReading, inTune], true)
    expect(out).toMatchObject({ done: 3, inTune: 1 })
  })

  it('counts a note taken twice once, and in tune only if both readings are', () => {
    const bothGood = [
      row(60, { register: 'MMM', reeds: [reed(true)] }),
      row(60, { register: 'M2', reeds: [reed(true)] }),
    ]
    expect(sessionProgress(sess(), bothGood, true)).toMatchObject({ done: 1, inTune: 1 })

    const oneBad = [
      row(60, { register: 'MMM', reeds: [reed(true)] }),
      row(60, { register: 'M2', reeds: [reed(false)] }),
    ]
    expect(sessionProgress(sess(), oneBad, true)).toMatchObject({ done: 1, inTune: 0 })
  })
})
