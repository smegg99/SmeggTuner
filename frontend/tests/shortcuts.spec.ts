import { describe, expect, it } from 'vitest'
import { shortcutFor } from '~/utils/shortcuts'

// The reference tuner's keys, and the ones SmeggTuner already had, without stepping on a recording's transport.
const tune = { tune: true, file: false }

describe('the global keys', () => {
  it('freezes on Space with the microphone, and leaves Space to the transport with a file', () => {
    expect(shortcutFor(' ', tune)).toBe('freeze')
    expect(shortcutFor(' ', { tune: true, file: true })).toBeUndefined()
    expect(shortcutFor('h', { tune: true, file: true })).toBe('freeze')
  })

  it('steps the note with the arrows, takes on Enter and undoes on Backspace, in the tune room', () => {
    expect(shortcutFor('ArrowLeft', tune)).toBe('noteDown')
    expect(shortcutFor('ArrowRight', tune)).toBe('noteUp')
    expect(shortcutFor('Enter', tune)).toBe('takeNow')
    expect(shortcutFor('Backspace', tune)).toBe('undo')
  })

  // The workshop has a table and a chart that use these keys for themselves.
  it('leaves Space, Enter, Backspace and the arrows alone outside the tune room', () => {
    const shop = { tune: false, file: false }
    for (const key of [' ', 'Enter', 'Backspace', 'ArrowLeft', 'ArrowRight']) {
      expect(shortcutFor(key, shop)).toBeUndefined()
    }
  })

  it('keeps Z for undo and the letter keys everywhere, whatever the case', () => {
    expect(shortcutFor('Z', { tune: false, file: false })).toBe('undo')
    expect(shortcutFor('R', { tune: false, file: false })).toBe('record')
    expect(shortcutFor('q', tune)).toBeUndefined()
  })
})
