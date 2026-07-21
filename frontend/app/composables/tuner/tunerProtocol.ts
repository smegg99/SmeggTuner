// core/tuning.NumNotes: one band per tracked note, E0..C9.
export const NOTE_BANDS = 105

// core/dsp.EqualizerCeilingDB: the range the wire packs a band against.
export const EQUALIZER_CEILING_DB = 60

// core/dsp.EqualizerFullScaleDB: what the strip draws against, not the ceiling above; confusing the two caps the tallest bar at 47%.
export const EQUALIZER_FULL_SCALE_DB = 27.9588

export const NOTE_MIN = 16
export const NOTE_MAX = 120
export const AUTO_NOTE = 0

// core/dsp.SpectrumColumns / SpectrumCents: the band arrives pre-binned, heights 0..1.
export const SPECTRUM_COLUMNS = 256
export const SPECTRUM_CENTS = 50

// core/dsp.SpectrumFloorDB: the decibel depth the bottom of the spectrum panel means.
export const SPECTRUM_FLOOR_DB = 60

// core/dsp.WaveformPoints.
export const WAVE_POINTS = 256

export const ERROR_DEVICE_LOST = 'tuner.error.deviceLost'

// Backend errors are i18n keys (services/audio/errors.go); anything else is a bug.
const ERROR_KEY_PATTERN = /tuner\.error\.[A-Za-z0-9]+/
const ERROR_UNEXPECTED = 'tuner.error.unexpected'

export type EngineState = 'initializing' | 'running' | 'frozen' | 'tooLoud' | 'tooQuiet' | 'deviceLost' | 'stopped'
export type TuningUnit = 'cent' | 'hz'

const ENGINE_STATES: readonly string[] = [
  'initializing', 'running', 'frozen', 'tooLoud', 'tooQuiet', 'deviceLost', 'stopped',
]

// toErrorKey pulls the i18n key out of a Wails rejection's Go error message.
export function toErrorKey(err: unknown): string {
  const message = err instanceof Error ? err.message : String(err)
  return ERROR_KEY_PATTERN.exec(message)?.[0] ?? ERROR_UNEXPECTED
}

export function toEngineState(raw: string, fallback: EngineState): EngineState {
  return ENGINE_STATES.includes(raw) ? raw as EngineState : fallback
}

// unpack inverts services/tuner.pack (v = byte / 255 * span); pictures ship as base64 bytes, not JSON floats, to stay cheap.
export function unpack(dst: Float32Array, src: string | undefined, span: number): number {
  if (!src) return 0
  const bin = atob(src)
  const n = Math.min(dst.length, bin.length)
  for (let i = 0; i < n; i++) dst[i] = (bin.charCodeAt(i) / 255) * span
  return n
}

// unpackSigned inverts services/tuner.packSigned; zero is 128 (centre line).
export function unpackSigned(dst: Float32Array, src: string | undefined): number {
  if (!src) return 0
  const bin = atob(src)
  const n = Math.min(dst.length, bin.length)
  for (let i = 0; i < n; i++) dst[i] = (bin.charCodeAt(i) - 128) / 127
  return n
}
