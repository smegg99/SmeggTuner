import type { Ref } from 'vue'
import type { MeasurementDTO, StateDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/models.js'
import { FrameKind } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { BeatMeasure, Lock, Needle, Panel, ReedMeasure } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import type { EngineState } from './tunerProtocol'
import { AUTO_NOTE, BAND_SPECTRUM_COLUMNS, EQUALIZER_CEILING_DB, ERROR_DEVICE_LOST, MAX_SPECTRA, toEngineState, unpack, unpackSigned } from './tunerProtocol'

// Hot draw buffers a canvas reads directly, kept outside Vue.
export interface LiveBuffers {
  spectrum: Float32Array
  hasSpectrum: boolean
  /** one picture per chord tone or register band; only the first spectraCount are current */
  spectra: Float32Array[]
  spectraCount: number
  bands: Float32Array
  wave: Float32Array
  waveLength: number
}

export interface TunerState {
  note: Ref<number>
  noteName: Ref<string>
  scalePitch: Ref<number>
  /** the chord name in chords mode, else empty */
  label: Ref<string>
  reeds: Ref<ReedMeasure[]>
  beats: Ref<BeatMeasure[]>
  reedErrors: Ref<ReedError[]>
  beatErrors: Ref<BeatError[]>
  panels: Ref<Panel[]>
  needles: Ref<Needle[]>
  needlesDimmed: Ref<boolean>
  lock: Ref<Lock>
  inputLevel: Ref<number>
  /** the capture lost audio in the last 0.1 s */
  unreliable: Ref<boolean>
  /** flips on every analysis, so a warning can show on alternate ones */
  blink: Ref<boolean>
  state: Ref<EngineState>
  running: Ref<boolean>
  readingAt: Ref<number | null>
  starting: Ref<boolean>
  error: Ref<string>
  live: LiveBuffers
  notify: () => void
}

export const NO_LOCK: Lock = { visible: false, internal: false, acquired: false, countdown: 0, recorded: false }

// Heartbeats repeat the screen state twenty times a second; replacing an equal value would re-render for nothing.
function same<T extends object>(a: readonly T[], b: readonly T[]): boolean {
  return a.length === b.length && a.every((x, i) => sameFields(x, b[i]!))
}

function sameFields<T extends object>(a: T, b: T): boolean {
  const keys = Object.keys(a) as (keyof T)[]
  return keys.length === Object.keys(b).length && keys.every(k => a[k] === b[k])
}

export function onMeasurement(m: MeasurementDTO, s: TunerState) {
  // A measurement that outlived its run would repaint a stopped engine's meters.
  if (!s.running.value) return

  s.inputLevel.value = m.inputLevel
  s.state.value = toEngineState(m.state, s.state.value)
  if (s.unreliable.value !== m.unreliable) s.unreliable.value = m.unreliable
  if (m.frame !== FrameKind.FrameNone) s.blink.value = !s.blink.value
  s.live.waveLength = unpackSigned(s.live.wave, m.waveform)

  // Every frame, heartbeat or analysis, carries the reference screen state.
  const panels = m.panels ?? []
  const needles = m.needles ?? []
  if (!same(panels, s.panels.value)) s.panels.value = panels
  if (!same(needles, s.needles.value)) s.needles.value = needles
  s.needlesDimmed.value = m.needlesDimmed
  if (!sameFields(m.lock, s.lock.value)) s.lock.value = m.lock

  // Only a measured analysis is a new reading; a heartbeat, a hold or an ineligible frame keeps the last.
  if (m.frame === FrameKind.FrameMeasured) {
    // readingAt is where in the recording this was measured; null for a live mic or sourceAt 0. core/dsp: Measurement.SourceAt.
    s.readingAt.value = m.sourceAt > 0 ? m.sourceAt : null
    s.note.value = m.note
    s.noteName.value = m.noteName
    s.scalePitch.value = m.scalePitch
    s.label.value = m.label ?? ''
    s.live.hasSpectrum = unpack(s.live.spectrum, m.spectrum, 1) > 0
    copySpectra(s.live, m.spectra)
    unpack(s.live.bands, m.equalizer, EQUALIZER_CEILING_DB)
    s.reeds.value = m.reeds ?? []
    s.beats.value = m.beats ?? []
    s.reedErrors.value = m.reedErrors ?? []
    s.beatErrors.value = m.beatErrors ?? []
  }

  s.notify()
}

// copySpectra keeps the per-tone pictures in reused buffers; a frame without them clears the count.
export function copySpectra(live: LiveBuffers, spectra: number[][] | undefined) {
  const n = Math.min(spectra?.length ?? 0, MAX_SPECTRA)
  for (let i = 0; i < n; i++) {
    const dst = live.spectra[i] ?? (live.spectra[i] = new Float32Array(BAND_SPECTRUM_COLUMNS))
    const src = spectra![i] ?? []
    for (let c = 0; c < BAND_SPECTRUM_COLUMNS; c++) dst[c] = src[c] ?? 0
  }
  live.spectraCount = n
}

export function clearLive(s: TunerState) {
  s.live.spectrum.fill(0)
  s.live.spectraCount = 0
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
  s.scalePitch.value = 0
  s.label.value = ''
  s.reeds.value = []
  s.beats.value = []
  s.reedErrors.value = []
  s.beatErrors.value = []
  s.panels.value = []
  s.needles.value = []
  s.needlesDimmed.value = true
  s.lock.value = NO_LOCK
  s.inputLevel.value = 0
  s.unreliable.value = false
  clearLive(s)
}
