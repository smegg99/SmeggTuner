import { describe, expect, it } from 'vitest'
import { a4Refusal, parseA4, stepTranspose, transposeRefusal } from '~/utils/reference'

const free = { lockA4: false, session: false, readings: 0, mode: 'tones' }

describe('the A4 control', () => {
  it('is refused by the lock, and by a session that holds readings', () => {
    expect(a4Refusal(free)).toBe('')
    expect(a4Refusal({ ...free, lockA4: true })).toBe('tuner.reference.a4Locked')
    expect(a4Refusal({ ...free, session: true, readings: 3 })).toBe('tuner.reference.frozen')
    // Without a session there are no readings to protect.
    expect(a4Refusal({ ...free, readings: 3 })).toBe('')
  })
})

describe('a typed A4', () => {
  it('reads a point or a comma onto the 0.1 Hz grid', () => {
    expect(parseA4('442')).toBe(442)
    expect(parseA4(' 441,26 ')).toBe(441.3)
  })

  it('refuses anything outside 430..450, and anything not a number', () => {
    expect(parseA4('429.9')).toBeNull()
    expect(parseA4('450.1')).toBeNull()
    expect(parseA4('')).toBeNull()
    expect(parseA4('A4')).toBeNull()
  })
})

describe('the transposition', () => {
  it('runs 0..11', () => {
    expect(stepTranspose(0, -1)).toBe(0)
    expect(stepTranspose(11, 1)).toBe(11)
    expect(stepTranspose(2, 1)).toBe(3)
  })

  it('is refused in chords and octaves, and once a session holds readings', () => {
    expect(transposeRefusal(free)).toBe('')
    expect(transposeRefusal({ ...free, mode: 'chords' })).toBe('tuner.reference.transposeMode')
    expect(transposeRefusal({ ...free, mode: 'octaves' })).toBe('tuner.reference.transposeMode')
    // A session's register decides, so the panel mode does not matter there.
    expect(transposeRefusal({ ...free, session: true, mode: 'chords' })).toBe('')
    expect(transposeRefusal({ ...free, session: true, readings: 1 })).toBe('tuner.reference.frozen')
  })
})
