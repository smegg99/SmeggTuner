import { describe, expect, it } from 'vitest'
import { panelBoxes, panelSlotIds, panelSlotOctaves } from '~/utils/panel'

// Without a session the panel decides: tones toggle reeds 1..3, octaves the boxes above a fixed base.
describe('the reed boxes', () => {
  it('toggles any reed in tones, and will not clear the last one', () => {
    const boxes = panelBoxes('tones', [3], [12])
    expect(boxes.map(b => b.on)).toEqual([false, false, true])
    expect(boxes[2]!.last).toBe(true)
    expect(boxes.some(b => b.fixed)).toBe(false)
  })

  it('keeps the base of octaves on and lets one upper box stay', () => {
    const boxes = panelBoxes('octaves', [1], [24])
    expect(boxes[0]).toEqual({ box: 1, on: true, fixed: true, last: false })
    expect(boxes[1]!.on).toBe(false)
    expect(boxes[2]).toMatchObject({ on: true, last: true })
  })

  it('has no boxes in chords', () => {
    expect(panelBoxes('chords', [1, 2], [12])).toEqual([])
  })
})

// The slots the engine measures, as dsp.ReedMeasure.Reed numbers them, and how far above the key each sounds.
describe('the measured slots', () => {
  it('numbers a tones gap by its boxes', () => {
    expect(panelSlotIds('tones', [3, 1], [])).toEqual([1, 3])
    expect(panelSlotOctaves('tones', [3, 1], [])).toEqual([0, 0])
  })

  it('numbers the octaves boxes 1, 2, 3 by how many octaves up they are', () => {
    expect(panelSlotIds('octaves', [1], [24, 12])).toEqual([1, 2, 3])
    expect(panelSlotIds('octaves', [1], [24])).toEqual([1, 3])
    expect(panelSlotOctaves('octaves', [1], [24])).toEqual([0, 24])
  })

  it('measures three chord tones at the key', () => {
    expect(panelSlotIds('chords', [1], [12])).toEqual([1, 2, 3])
    expect(panelSlotOctaves('chords', [1], [12])).toEqual([0, 0, 0])
  })
})
