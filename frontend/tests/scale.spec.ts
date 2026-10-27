import { describe, expect, it } from 'vitest'
import { EQUALIZER_CEILING_DB, EQUALIZER_FULL_SCALE_DB } from '~/composables/useTuner'
import { RULER_CENTS, bandToFrac, centToFrac } from '~/utils/scale'

describe('the ruler scale', () => {
  it('runs fifty units either side, about a semitone end to end', () => {
    expect(RULER_CENTS).toBe(50)
  })

  it('puts the scale pitch dead centre', () => {
    expect(centToFrac(0)).toBeCloseTo(0.5)
  })

  it('puts the ends at the ends', () => {
    expect(centToFrac(-50)).toBeCloseTo(0)
    expect(centToFrac(50)).toBeCloseTo(1)
  })

  // A 16' reed sounds an octave below the key, and a badly out reed can read
  // hundreds of cents off. Drawn unclamped that is a needle painted outside the
  // canvas, which is a needle the technician cannot see and does not know is there.
  it('clamps a reed that is off the scale onto the edge, not off the canvas', () => {
    expect(centToFrac(-400)).toBe(0)
    expect(centToFrac(400)).toBe(1)
  })
})

describe('the note strip scale', () => {
  // core/dsp: each band is the note's coarse score over the learned background in dB, clamped to
  // the ceiling, and EqualizerFullScaleDB is that ceiling.
  it('draws against core/dsp\'s full scale, which is the wire ceiling', () => {
    expect(EQUALIZER_FULL_SCALE_DB).toBe(EQUALIZER_CEILING_DB)
  })

  it('puts the sounding note at the top of the strip, not at 47% of it', () => {
    expect(bandToFrac(EQUALIZER_FULL_SCALE_DB)).toBe(1)
  })

  // A silent room now arrives as zeroes rather than as a confident little mountain
  // range: core/dsp's TestTheEqualizerStaysDarkInSilence.
  it('draws nothing for a band with nothing in it', () => {
    expect(bandToFrac(0)).toBe(0)
  })

  it('keeps the shape of the sound in between', () => {
    expect(bandToFrac(EQUALIZER_FULL_SCALE_DB / 2)).toBeCloseTo(0.5)
    expect(bandToFrac(EQUALIZER_FULL_SCALE_DB / 4)).toBeCloseTo(0.25)
  })

  // a byte off the wire is not a promise
  it('clamps rather than drawing outside its panel', () => {
    expect(bandToFrac(EQUALIZER_CEILING_DB)).toBe(1)
    expect(bandToFrac(-5)).toBe(0)
  })
})
