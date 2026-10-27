import { describe, expect, it } from 'vitest'
import { NEEDLE_GRAB_PX, outsideRange, pressKind } from '~/utils/waveGesture'

describe('wave gesture', () => {
  it('scrubs from the timeline lane wherever the playhead is', () => {
    expect(pressKind(400, 100, true)).toBe('scrub')
    expect(pressKind(400, null, true)).toBe('scrub')
  })

  it('grabs the playhead within a few pixels of it', () => {
    expect(pressKind(100 + NEEDLE_GRAB_PX, 100, false)).toBe('scrub')
    expect(pressKind(100 - NEEDLE_GRAB_PX, 100, false)).toBe('scrub')
    expect(pressKind(100 + NEEDLE_GRAB_PX + 1, 100, false)).toBe('press')
  })

  it('treats a press away from the playhead as a click or a range drag', () => {
    expect(pressKind(400, 100, false)).toBe('press')
    expect(pressKind(400, null, false)).toBe('press')
  })

  it('knows when a seek would be clamped into the selected range', () => {
    expect(outsideRange(3, 5, 7)).toBe(true)
    expect(outsideRange(8, 5, 7)).toBe(true)
    expect(outsideRange(6, 5, 7)).toBe(false)
    expect(outsideRange(5, 5, 7)).toBe(false)
  })
})
