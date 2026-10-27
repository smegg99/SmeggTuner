import { describe, expect, it, vi } from 'vitest'
import { ref, shallowRef } from 'vue'
import type { Lock, Needle, Panel } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { MeasurementDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/models.js'
import { BAND_SPECTRUM_COLUMNS, NOTE_BANDS, SPECTRUM_COLUMNS, WAVE_POINTS } from '~/composables/tuner/tunerProtocol'
import { NO_LOCK, clearReading, onMeasurement } from '~/composables/tuner/tunerReducers'
import type { TunerState } from '~/composables/tuner/tunerReducers'

/*
 * THE REDUCER KNOWS A HEARTBEAT FROM A READING.
 *
 * The engine analyses twice a second and sends a heartbeat twenty times a second between. Every
 * frame carries the reference screen state (readouts, needles, lock) and the meters; only a
 * measured analysis carries a new reading. A heartbeat taken for a reading would blank the
 * reeds, and a held or ineligible frame taken for one would replace the reading the screen holds.
 */

function state(): TunerState {
  return {
    note: ref(0),
    noteName: ref(''),
    scalePitch: ref(0),
    label: ref(''),
    reeds: shallowRef([]),
    beats: shallowRef([]),
    reedErrors: shallowRef([]),
    beatErrors: shallowRef([]),
    panels: shallowRef([]),
    needles: shallowRef([]),
    needlesDimmed: ref(true),
    lock: shallowRef(NO_LOCK),
    inputLevel: ref(0),
    unreliable: ref(false),
    blink: ref(false),
    state: ref('running'),
    running: ref(true),
    readingAt: ref(null),
    starting: ref(false),
    error: ref(''),
    live: {
      spectrum: new Float32Array(SPECTRUM_COLUMNS),
      hasSpectrum: false,
      spectra: [],
      spectraCount: 0,
      bands: new Float32Array(NOTE_BANDS),
      wave: new Float32Array(WAVE_POINTS),
      waveLength: 0,
    },
    notify: vi.fn(),
  }
}

const panels = (value: number): Panel[] => [{ value, target: -99, signed: true, decimals: 1, dimmed: false, shown: `+${value}` }]
const needles = (dev: number): Needle[] => [{ dev, freq: 440, valid: true }]
const locked: Lock = { visible: true, internal: true, acquired: true, countdown: 4, recorded: false }
const bytes = (n: number, v: number) => btoa(String.fromCharCode(...Array.from({ length: n }, () => v)))

function frame(opts: Partial<MeasurementDTO>): MeasurementDTO {
  return {
    state: 'running',
    frame: '',
    note: 0,
    noteName: '',
    scalePitch: 0,
    reeds: [],
    beats: [],
    panels: panels(1),
    needles: needles(1),
    needlesDimmed: false,
    lock: NO_LOCK,
    learning: false,
    inputLevel: 0.3,
    sourceAt: 0,
    equalizer: '',
    spectrum: '',
    waveform: bytes(WAVE_POINTS, 200),
    reedErrors: [],
    beatErrors: [],
    ...opts,
  } as MeasurementDTO
}

const measured = (opts: Partial<MeasurementDTO> = {}) => frame({
  frame: 'measured' as MeasurementDTO['frame'],
  note: 69,
  noteName: 'A4',
  scalePitch: 440,
  reeds: [{ present: true, freq: 440.5, dev: 1.9 }],
  reedErrors: [{ reed: 0, curr: 1.9, goal: 0, error: 1.9, currHz: 0.5, goalHz: 0, errorHz: 0.5, inTol: false }],
  spectrum: bytes(SPECTRUM_COLUMNS, 255),
  equalizer: bytes(NOTE_BANDS, 128),
  sourceAt: 2.5,
  panels: panels(1.9),
  needles: needles(1.9),
  lock: locked,
  ...opts,
})

describe('the tuner reducer', () => {
  it('takes a measured analysis as the reading', () => {
    const s = state()
    onMeasurement(measured(), s)

    expect(s.note.value).toBe(69)
    expect(s.noteName.value).toBe('A4')
    expect(s.scalePitch.value).toBe(440)
    expect(s.reeds.value).toHaveLength(1)
    expect(s.reedErrors.value[0]?.error).toBeCloseTo(1.9)
    expect(s.readingAt.value).toBe(2.5)
    expect(s.live.hasSpectrum).toBe(true)
    expect(s.live.bands[0]).toBeGreaterThan(0)
    expect(s.panels.value[0]?.value).toBeCloseTo(1.9)
    expect(s.lock.value.visible).toBe(true)
  })

  it('takes a heartbeat as the meters and the screen state, never as a reading', () => {
    const s = state()
    onMeasurement(measured(), s)
    onMeasurement(frame({ inputLevel: 0.7, panels: panels(2.4), needles: needles(2.4), needlesDimmed: true, lock: NO_LOCK }), s)

    expect(s.inputLevel.value).toBeCloseTo(0.7)
    expect(s.live.waveLength).toBe(WAVE_POINTS)
    expect(s.panels.value[0]?.value).toBeCloseTo(2.4)
    expect(s.needles.value[0]?.dev).toBeCloseTo(2.4)
    expect(s.needlesDimmed.value).toBe(true)
    expect(s.lock.value.visible).toBe(false)

    // The reading stays exactly as the last analysis left it.
    expect(s.note.value).toBe(69)
    expect(s.reeds.value).toHaveLength(1)
    expect(s.reedErrors.value).toHaveLength(1)
    expect(s.readingAt.value).toBe(2.5)
    expect(s.live.hasSpectrum).toBe(true)
  })

  // A held result (stop after lock) and an ineligible one leave the screen's reading alone.
  it('keeps the reading through held and ineligible analyses', () => {
    const s = state()
    onMeasurement(measured(), s)
    for (const kind of ['held', 'ineligible', 'noNote', 'waiting']) {
      onMeasurement(frame({ frame: kind as MeasurementDTO['frame'], note: 70, scalePitch: 466, reeds: [{ present: true, freq: 470, dev: 30 }] }), s)
    }
    expect(s.note.value).toBe(69)
    expect(s.reeds.value[0]?.dev).toBeCloseTo(1.9)
  })

  // Twenty identical screen states a second must not re-render the boxes twenty times.
  it('keeps the same readouts when a heartbeat repeats them', () => {
    const s = state()
    onMeasurement(measured(), s)
    const before = s.panels.value
    const lock = s.lock.value
    onMeasurement(frame({ panels: panels(1.9), needles: needles(1.9), lock: { ...locked } }), s)
    expect(s.panels.value).toBe(before)
    expect(s.lock.value).toBe(lock)
  })

  it('ignores a measurement that outlived its run', () => {
    const s = state()
    s.running.value = false
    onMeasurement(measured(), s)
    expect(s.note.value).toBe(0)
    expect(s.panels.value).toEqual([])
  })

  it('wipes the screen state with the reading', () => {
    const s = state()
    onMeasurement(measured(), s)
    clearReading(s)
    expect(s.panels.value).toEqual([])
    expect(s.needles.value).toEqual([])
    expect(s.lock.value.visible).toBe(false)
    expect(s.reeds.value).toEqual([])
    expect(s.note.value).toBe(0)
  })

  // Chords name the chord, and a chord or a register spanning octaves brings one picture per tone.
  it('takes the chord name and the per-tone pictures with the reading', () => {
    const s = state()
    const tone = (v: number) => Array.from({ length: BAND_SPECTRUM_COLUMNS }, () => v)
    onMeasurement(measured({ label: 'Cm', spectra: [tone(0.1), tone(0.2), tone(0.3)] }), s)
    expect(s.label.value).toBe('Cm')
    expect(s.live.spectraCount).toBe(3)
    expect(s.live.spectra[2]?.[10]).toBeCloseTo(0.3)

    // A single-band reading carries no pictures: the count drops, the buffers stay for reuse.
    onMeasurement(measured(), s)
    expect(s.label.value).toBe('')
    expect(s.live.spectraCount).toBe(0)
  })

  // The clip warning shows on alternate analyses; heartbeats between them do not move it.
  it('flips the blink on every analysis and never on a heartbeat', () => {
    const s = state()
    onMeasurement(measured(), s)
    const after = s.blink.value
    onMeasurement(frame({}), s)
    expect(s.blink.value).toBe(after)
    onMeasurement(frame({ frame: 'noNote' as MeasurementDTO['frame'] }), s)
    expect(s.blink.value).toBe(!after)
  })

  it('follows the lost-audio flag on every frame and forgets it with the reading', () => {
    const s = state()
    onMeasurement(frame({ unreliable: true }), s)
    expect(s.unreliable.value).toBe(true)
    clearReading(s)
    expect(s.unreliable.value).toBe(false)
  })
})
