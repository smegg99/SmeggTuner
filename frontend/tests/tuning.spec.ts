import { describe, expect, it } from 'vitest'
import { noteName } from '~/utils/tuning'

/*
 * Every naming backend/common/config/config.cue allows.
 *
 * This exists because `polish` was MISSING and the miss was silent: it fell through
 * to the default and came back as sharps, so the note panel (which uses the engine's
 * own name) said Gis4 while the toolbar (which uses this table) said G#4. Two names
 * for one note, in one window, at the same time.
 *
 * A naming that is not in the switch must not quietly become sharps. It must fail
 * here first.
 */
const NAMINGS = ['cdefgab', 'cdefgah', 'doremi', 'polish'] as const

describe('noteName', () => {
  // MIDI 68 is G#4 / Gis4. The awkward one, and the one that exposed the bug.
  it('names the sharp of G in every scale the config allows', () => {
    expect(noteName(68, 'cdefgab')).toBe('G#4')
    expect(noteName(68, 'cdefgah')).toBe('G#4')
    expect(noteName(68, 'doremi')).toBe('Sol#4')
    expect(noteName(68, 'polish')).toBe('Gis4')
  })

  // The Polish and German scales part company at B: H is B natural, and B is B flat.
  it('spells the eleventh and twelfth degrees the way each scale does', () => {
    expect(noteName(70, 'cdefgab')).toBe('A#4')
    expect(noteName(71, 'cdefgab')).toBe('B4')

    expect(noteName(70, 'cdefgah')).toBe('B4')
    expect(noteName(71, 'cdefgah')).toBe('H4')

    expect(noteName(70, 'polish')).toBe('Ais4')
    expect(noteName(71, 'polish')).toBe('H4')
  })

  it('agrees on middle C, which every scale calls C4', () => {
    for (const naming of NAMINGS) {
      expect(noteName(60, naming)).toMatch(/^(C|Do)4$/)
    }
  })

  // No naming may silently fall back to another one: that is the exact shape of the
  // defect this file was written for.
  it('gives every allowed naming its own spelling of the black keys', () => {
    const spellings = NAMINGS.map(n => noteName(66, n))
    expect(new Set(spellings).size).toBe(NAMINGS.length - 1) // cdefgab and cdefgah share F#
    expect(noteName(66, 'polish')).toBe('Fis4')
  })
})
