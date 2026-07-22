import type { Ref } from 'vue'
import type { MeasurementDTO, StateDTO } from '~~bindings/smegg.me/smeggtuner/services/tuner/models.js'
import type { BandReport, BeatMeasure, ReedMeasure } from '~~bindings/smegg.me/smeggtuner/core/dsp/models.js'
import type { BeatError, ReedError } from '~/types/record'
import type { EngineState } from './tunerProtocol'
import { AUTO_NOTE, EQUALIZER_CEILING_DB, ERROR_DEVICE_LOST, toEngineState, unpack, unpackSigned } from './tunerProtocol'

// Hot draw buffers a canvas reads directly, kept outside Vue.
export interface LiveBuffers {
  spectrum: Float32Array
  hasSpectrum: boolean
  bands: Float32Array
  wave: Float32Array
  waveLength: number
}

export interface TunerState {
  note: Ref<number>
  noteName: Ref<string>
  locked: Ref<boolean>
  lockProgress: Ref<number>
  scalePitch: Ref<number>
  reeds: Ref<ReedMeasure[]>
  beats: Ref<BeatMeasure[]>
  bands: Ref<BandReport[]>
  reedsSeparated: Ref<boolean>
  reedsFromBeat: Ref<boolean>
  reedErrors: Ref<ReedError[]>
  beatErrors: Ref<BeatError[]>
  inputLevel: Ref<number>
  state: Ref<EngineState>
  running: Ref<boolean>
  readingAt: Ref<number | null>
  starting: Ref<boolean>
  error: Ref<string>
  live: LiveBuffers
  notify: () => void
}

// A measurement carries a reading only when it carries reeds; testing the spectrum instead would blank the rows mid-note.
function carriesFineResult(m: MeasurementDTO): boolean {
  return m.reeds.length > 0
}

export function onMeasurement(m: MeasurementDTO, s: TunerState) {
  // A measurement that outlived its run would repaint a stopped engine's meters.
  if (!s.running.value) return

  s.inputLevel.value = m.inputLevel
  s.state.value = toEngineState(m.state, s.state.value)
  // lockProgress is 0..1; the engine sends 0 on a heartbeat.
  s.lockProgress.value = m.lockProgress

  // The waveform trace updates on every measurement, whatever the engine's result.
  s.live.waveLength = unpackSigned(s.live.wave, m.waveform)

  // Spectrum, tracked note and strip are latched together on spectrum presence so bars and name can't contradict.
  if (m.spectrum.length > 0 && m.scalePitch > 0) {
    // readingAt is where in the recording this was measured; null for a live mic or sourceAt 0. core/dsp: Measurement.SourceAt.
    s.readingAt.value = m.sourceAt > 0 ? m.sourceAt : null

    s.note.value = m.note
    s.noteName.value = m.noteName
    s.scalePitch.value = m.scalePitch
    s.live.hasSpectrum = unpack(s.live.spectrum, m.spectrum, 1) > 0
    unpack(s.live.bands, m.equalizer, EQUALIZER_CEILING_DB)
  }

  // The reading is latched as a group; a flag or lock outliving its reeds would describe the previous note.
  if (carriesFineResult(m)) {
    s.locked.value = m.locked
    s.reeds.value = m.reeds
    s.beats.value = m.beats
    s.bands.value = m.bands ?? []
    s.reedsSeparated.value = m.reedsSeparated
    s.reedsFromBeat.value = m.reedsFromBeat
    s.reedErrors.value = m.reedErrors ?? []
    s.beatErrors.value = m.beatErrors ?? []
  }

  s.notify()
}

export function clearLive(s: TunerState) {
  s.live.spectrum.fill(0)
  s.live.bands.fill(0)
  s.live.wave.fill(0)
  s.live.hasSpectrum = false
  s.live.waveLength = 0
  s.notify()
}

export function onState(dto: StateDTO, s: TunerState) {
  s.running.value = dto.running
  s.error.value = dto.error

  if (dto.running) return

  // A stopped engine keeps its last picture; only input level blanks. Device-lost is reported here, never on the measurement stream.
  s.state.value = s.error.value === ERROR_DEVICE_LOST ? 'deviceLost' : 'stopped'
  s.inputLevel.value = 0
}

// clearReading wipes all reading state; runs at every Start and source change (a stale reading against a new recording looks plausible and is wrong).
export function clearReading(s: TunerState) {
  s.readingAt.value = null
  s.note.value = AUTO_NOTE
  s.noteName.value = ''
  s.locked.value = false
  s.lockProgress.value = 0
  s.scalePitch.value = 0
  s.reeds.value = []
  s.beats.value = []
  s.bands.value = []
  s.reedsSeparated.value = true
  s.reedsFromBeat.value = false
  s.reedErrors.value = []
  s.beatErrors.value = []
  s.inputLevel.value = 0
  clearLive(s)
}
