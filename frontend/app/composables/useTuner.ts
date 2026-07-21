// Detached singleton for the tuner: reactive reading, hot draw buffers and engine subscriptions. The hot arrays never enter Vue; they land in `live` ~12/s and canvases subscribe via onData.
import { computed, effectScope, ref, shallowRef, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/services/tuner/service.js'
import type { MeasurementDTO, StateDTO, SettingsDTO } from '~~bindings/smegg.me/smeggtuner/services/tuner/models.js'
import type { BeatMeasure, ReedMeasure } from '~~bindings/smegg.me/smeggtuner/core/dsp/models.js'
import type { Config } from '~~bindings/smegg.me/smeggtuner/common/config/models.js'
import type { BeatError, ReedError } from '~/types/record'
import { AUTO_NOTE, EQUALIZER_CEILING_DB, EQUALIZER_FULL_SCALE_DB, NOTE_BANDS, NOTE_MAX, NOTE_MIN, SPECTRUM_CENTS, SPECTRUM_COLUMNS, SPECTRUM_FLOOR_DB, WAVE_POINTS, toErrorKey } from './tuner/tunerProtocol'
import type { EngineState, TuningUnit } from './tuner/tunerProtocol'
import { clearReading as clearReadingImpl, onMeasurement, onState } from './tuner/tunerReducers'
import type { LiveBuffers, TunerState } from './tuner/tunerReducers'
import { apply, playTone, startEngine, stopEngine, stopTone } from './tuner/tunerCommands'

export { AUTO_NOTE, EQUALIZER_CEILING_DB, EQUALIZER_FULL_SCALE_DB, NOTE_BANDS, NOTE_MAX, NOTE_MIN, SPECTRUM_CENTS, SPECTRUM_COLUMNS, SPECTRUM_FLOOR_DB, WAVE_POINTS, toErrorKey }
export type { EngineState, TuningUnit }

const EVENT_MEASUREMENT = 'tuner:measurement'
const EVENT_STATE = 'tuner:state'
const EVENT_SETTINGS = 'tuner:settings'

// Hot data, outside Vue: read it in a draw call; re-rendering a component on it is wrong.
export const live: LiveBuffers = {
  spectrum: new Float32Array(SPECTRUM_COLUMNS),
  hasSpectrum: false,
  bands: new Float32Array(NOTE_BANDS),
  wave: new Float32Array(WAVE_POINTS),
  waveLength: 0,
}

type Listener = () => void
const listeners = new Set<Listener>()

// onData fires once per measurement, after `live` is written.
export function onData(fn: Listener) {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

// The reading, latched as a group with the reeds it belongs to.
const note = ref(AUTO_NOTE)
const noteName = ref('')
const locked = ref(false)
// The engine's settle toward a lock, 0..1; feeds the calibration capture bar.
const lockProgress = ref(0)
const scalePitch = ref(0)
const reeds = shallowRef<ReedMeasure[]>([])
const beats = shallowRef<BeatMeasure[]>([])
const reedsSeparated = ref(true)
const reedsFromBeat = ref(false)

// The reeds/beats scored against the session's goal curve; no session = plain scale deviation.
const reedErrors = shallowRef<ReedError[]>([])
const beatErrors = shallowRef<BeatError[]>([])

const inputLevel = ref(0)
const state = ref<EngineState>('stopped')
const running = ref(false)

// Where in the recording the on-screen reading was measured, or null when there is no answer.
const readingAt = ref<number | null>(null)
const starting = ref(false)
const error = ref('')

const a4 = ref(0)
// Reeds the engine resolves per note: the pulled register's bank count, kept live by EVENT_SETTINGS.
const reedCount = ref(1)
const manualNote = ref(AUTO_NOTE)
const transpose = ref(0)
const frozen = ref(false)

// Freeze needs a reading to hold: while running, or after a stop left its last note.
const canFreeze = computed(() => running.value || note.value !== AUTO_NOTE)

// unit is display-only and lives in the config; the holder covers its async arrival.
const configRef = shallowRef<Config | null>(null)

const unit = computed<TuningUnit>({
  get: () => (configRef.value?.tuner.unit === 'hz' ? 'hz' : 'cent'),
  set: (next) => {
    const config = configRef.value
    if (config) config.tuner.unit = next
  },
})

const tunerState: TunerState = {
  note, noteName, locked, lockProgress, scalePitch, reeds, beats, reedsSeparated, reedsFromBeat,
  reedErrors, beatErrors, inputLevel, state, running, readingAt, starting, error,
  live,
  notify: () => { for (const fn of listeners) fn() },
}

// Wipe the screen. Called at every change of source: see clearReading.
const clearReading = () => clearReadingImpl(tunerState)

const scope = effectScope(true)
let started = false
let seeded = false

type Logger = ReturnType<typeof useLogger>

function start(log: Logger, config: Config) {
  configRef.value = config

  if (started) return
  started = true

  scope.run(() => {
    const offMeasurement = Events.On(EVENT_MEASUREMENT, (ev: { data: MeasurementDTO }) => {
      onMeasurement(ev.data, tunerState)
    })
    const offState = Events.On(EVENT_STATE, (ev: { data: StateDTO }) => {
      onState(ev.data, tunerState)
    })
    const offSettings = Events.On(EVENT_SETTINGS, (ev: { data: SettingsDTO }) => {
      reedCount.value = ev.data.reedCount
    })

    // Detached scope: without this a hot reload leaks one listener set per edit.
    if (import.meta.hot) {
      import.meta.hot.dispose(() => {
        offMeasurement()
        offState()
        offSettings()
        scope.stop()
        started = false
      })
    }

    // Config loads async: seed the A4 control once, then the control owns it.
    watch(() => config.tuner, (tuner) => {
      if (seeded || !tuner.a4) return
      seeded = true
      a4.value = tuner.a4
    }, { immediate: true, deep: true })
  })

  // Catch a count the engine resolved before this listener was up.
  TunerService.Settings()
    .then((s) => { reedCount.value = s.reedCount })
    .catch(() => {})

  TunerService.IsRunning()
    .then((on) => {
      running.value = on
      if (on) state.value = 'running'
    })
    .catch((err) => {
      log.error('tuner: failed to read engine state', { error: String(err) })
    })
}

export function useTuner() {
  const log = useLogger()
  const { config } = useConfigSync()
  start(log, config)

  return {
    note,
    noteName,
    locked,
    lockProgress,
    scalePitch,
    reeds,
    beats,
    reedsSeparated,
    reedsFromBeat,
    reedErrors,
    beatErrors,
    inputLevel,
    state,

    running,
    starting,
    readingAt,
    clearReading,
    error,
    start: () => startEngine(tunerState, log),
    stop: () => stopEngine(tunerState, log),

    a4,
    reedCount,
    manualNote,
    transpose,
    frozen,
    canFreeze,
    unit,
    setA4: (hz: number) => apply(tunerState, a4, hz, () => TunerService.SetA4(hz), log, 'a4'),
    setManualNote: (n: number) => apply(tunerState, manualNote, n, () => TunerService.SetManualNote(n), log, 'manual note'),
    setTranspose: (semitones: number) => apply(tunerState, transpose, semitones, () => TunerService.SetTranspose(semitones), log, 'transpose'),
    setFrozen: (on: boolean) => apply(tunerState, frozen, on, () => TunerService.Freeze(on), log, 'freeze'),
    playTone: (n: number) => playTone(tunerState, n, log),
    stopTone: () => stopTone(log),
  }
}
