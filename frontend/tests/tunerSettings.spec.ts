import { describe, expect, it } from 'vitest'
import type { SettingsDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/models.js'
import { applySettings, toMode, tunerSettings } from '~/composables/tuner/tunerSettings'

const dto = (o: Partial<SettingsDTO>): SettingsDTO => ({
  a4: 441.5, reedCount: 2, sessionReeds: 0, tolerance: 1, beatTolerance: 3,
  mode: 'tones', reedBoxes: [3, 1], octaveOffsets: [12], transpose: 2, manualNote: 69, ...o,
}) as SettingsDTO

// tuner:settings is the service's word on what the engine measures with; the screen never derives it.
describe('following the tuner settings', () => {
  it('takes every field, with the boxes in order', () => {
    const s = tunerSettings()
    applySettings(dto({}), s)
    expect(s.a4.value).toBe(441.5)
    expect(s.reedBoxes.value).toEqual([1, 3])
    expect(s.transpose.value).toBe(2)
    expect(s.manualNote.value).toBe(69)
  })

  it('keeps the same list when the boxes did not change, so nothing re-renders', () => {
    const s = tunerSettings()
    applySettings(dto({}), s)
    const boxes = s.reedBoxes.value
    applySettings(dto({ a4: 440 }), s)
    expect(s.reedBoxes.value).toBe(boxes)
  })

  it('reads an unknown mode as tones', () => {
    expect(toMode('chords')).toBe('chords')
    expect(toMode('')).toBe('tones')
  })
})
