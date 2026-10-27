import { describe, expect, it } from 'vitest'
import type { GoalRowDTO, RecordedDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'
import { anchorUnit, curveBounds, goalIn, goalShown } from '~/utils/curveUnits'
import { curveRange, curveSeries, noteTicks } from '~/utils/curveSeries'
import { toCents } from '~/utils/deviation'

// A goal in Hz is the reference beat list's rule: the note's pitch at A4 = 440 Hz, times the widened 0.06.
describe('restating a goal', () => {
  it('reads Hz at the note for A4 = 440, with the binary32 0.06', () => {
    expect(goalIn(10, 69, 'hz')).toBeCloseTo(10 * 440 * Math.fround(0.06) / 100, 12)
    expect(goalIn(10, 81, 'hz')).toBeCloseTo(2 * goalIn(10, 69, 'hz'), 12)
  })

  it('leaves reference units alone and converts cents', () => {
    expect(goalIn(7, 60, 'reference')).toBe(7)
    expect(goalIn(7, 60, 'cents')).toBeCloseTo(toCents(7))
  })

  it('rounds Hz to hundredths and the rest to tenths', () => {
    expect(goalShown(10, 69, 'hz')).toBe(2.64)
    expect(goalShown(7.04, 60, 'reference')).toBe(7)
  })

  it('tells SetAnchor the unit in core/target spelling', () => {
    expect(anchorUnit('hz')).toBe('hz')
    expect(anchorUnit('cents')).toBe('cent')
    expect(anchorUnit('reference')).toBe('reference')
  })
})

describe('the graph scale', () => {
  it('is symmetric when fixed', () => {
    expect(curveBounds([30], 10, 'reference')).toEqual({ min: -10, max: 10 })
  })

  it('fits outward with zero in when automatic', () => {
    expect(curveBounds([3, 8], 0, 'reference')).toEqual({ min: -5, max: 10 })
    expect(curveBounds([], 0, 'hz')).toEqual({ min: -0.5, max: 0.5 })
  })
})

const goal = (note: number, reeds: (number | null)[]): GoalRowDTO => ({ note, reeds }) as GoalRowDTO

describe('the curve series', () => {
  const goals = [goal(60, [0, 5]), goal(61, [0, null]), goal(64, [0, 6])]
  const recorded = [{ note: 62, take: 0, reeds: [{ column: 1, dev: 4, goal: 5 }] }] as RecordedDTO[]

  it('breaks the goal line where a note has none and where the covered range skips notes', () => {
    const [, , upper] = curveSeries({ anchors: [], goals, recorded, ranks: 2, unit: 'reference', showRecorded: false })
    expect(upper!.kind).toBe('goal')
    expect(upper!.points).toEqual([{ x: 60, y: 5 }, { x: 61, y: null }, { x: 63.5, y: null }, { x: 64, y: 6 }])
  })

  it('lays each reading on its own rank, only when asked', () => {
    const off = curveSeries({ anchors: [], goals, recorded, ranks: 2, unit: 'reference', showRecorded: false })
    expect(off.some(s => s.kind === 'recorded')).toBe(false)
    const on = curveSeries({ anchors: [], goals, recorded, ranks: 2, unit: 'reference', showRecorded: true })
    const rec = on.filter(s => s.kind === 'recorded')
    expect(rec[0]!.points).toEqual([])
    expect(rec[1]!.points).toEqual([{ x: 62, y: 4 }])
  })

  it('spans what it plots, a note past each end', () => {
    const series = curveSeries({ anchors: [], goals, recorded, ranks: 2, unit: 'reference', showRecorded: true })
    expect(curveRange(series)).toEqual({ min: 59, max: 65 })
    expect(curveRange([])).toBeNull()
  })

  // Note names on the axis: every C across a keyboard, finer on a short stretch.
  it('ticks at every C when three fit, finer otherwise', () => {
    expect(noteTicks(52, 90)).toEqual([60, 72, 84])
    expect(noteTicks(57, 70)).toEqual([57, 60, 63, 66, 69])
    expect(noteTicks(59, 62)).toEqual([59, 60, 61, 62])
  })
})
