import { describe, expect, it } from 'vitest'
import { DARK, LIGHT, SEMANTIC } from '~/theme/tokens'

const HEX = /^#[0-9a-f]{6}$/i

describe('the theme tokens', () => {
  it('defines the same names in both themes', () => {
    expect(Object.keys(LIGHT).sort()).toEqual(Object.keys(DARK).sort())
  })

  it('is all hex, so a canvas can use it without parsing', () => {
    for (const value of [...Object.values(LIGHT), ...Object.values(DARK)]) {
      expect(value).toMatch(HEX)
    }
  })

  // The measurement hues, asserted: exactly three that carry a reading's meaning, and
  // no fourth creeping in among them. The accent is separate and deliberately not one
  // of these (see below).
  it('carries exactly three semantic hues', () => {
    expect(SEMANTIC).toEqual(['reed', 'goal', 'warn'])

    for (const name of SEMANTIC) {
      expect(LIGHT[name]).toMatch(HEX)
      expect(DARK[name]).toMatch(HEX)
    }
  })

  // The one accent: a single interactive hue, present in both themes and hex like the
  // rest, but held OUT of the semantic set so it can never be read as a measurement.
  it('has one accent, and it is not a semantic hue', () => {
    expect(Object.keys(LIGHT)).toContain('accent')
    expect(Object.keys(DARK)).toContain('accent')
    expect(LIGHT.accent).toMatch(HEX)
    expect(DARK.accent).toMatch(HEX)
    expect(SEMANTIC as readonly string[]).not.toContain('accent')
  })

  // Light is not an inverted dark: it is designed. If someone ever generates one
  // from the other, these stop differing in the way a designed pair does.
  it('does not invert one theme into the other', () => {
    expect(LIGHT.ink).not.toBe(DARK.well)
    expect(LIGHT.well).not.toBe(DARK.ink)
  })
})
