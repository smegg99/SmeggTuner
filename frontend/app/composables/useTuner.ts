// Detached singleton for the tuner: reactive reading, hot draw buffers and engine subscriptions. The hot arrays never enter Vue; they land in `live` ~20/s and canvases subscribe via onData.
import { computed, effectScope, ref, shallowRef } from 'vue'
import { Events } from '@wailsio/runtime'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js'
import type { MeasurementDTO, StateDTO, SettingsDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/models.js'
import type { BeatMeasure, Lock, Needle, Panel, ReedMeasure } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import { AUTO_NOTE, EQUALIZER_CEILING_DB, EQUALIZER_FULL_SCALE_DB, NOTE_BANDS, NOTE_MAX, NOTE_MIN, SPECTRUM_CENTS, SPECTRUM_COLUMNS, SPECTRUM_FLOOR_DB, WAVE_POINTS, toErrorKey } from './tuner/tunerProtocol'
import type { EngineState } from './tuner/tunerProtocol'
import { NO_LOCK, clearReading as clearReadingImpl, onMeasurement, onState } from './tuner/tunerReducers'
import type { LiveBuffers, TunerState } from './tuner/tunerReducers'
import { apply, command, playTone, startEngine, stopEngine, stopTone } from './tuner/tunerCommands'
import { applySettings, tunerSettings } from './tuner/tunerSettings'
import type { PanelMode } from '~/utils/panel'

export { AUTO_NOTE, EQUALIZER_CEILING_DB, EQUALIZER_FULL_SCALE_DB, NOTE_BANDS, NOTE_MAX, NOTE_MIN, SPECTRUM_CENTS, SPECTRUM_COLUMNS, SPECTRUM_FLOOR_DB, WAVE_POINTS, toErrorKey }
export type { EngineState }

const EVENT_MEASUREMENT = 'tuner:measurement'
const EVENT_STATE = 'tuner:state'
const EVENT_SETTINGS = 'tuner:settings'

// Hot data, outside Vue: read it in a draw call; re-rendering a component on it is wrong.
export const live: LiveBuffers = {
  spectrum: new Float32Array(SPECTRUM_COLUMNS),
  hasSpectrum: false,
  spectra: [],
  spectraCount: 0,
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

// The last measured analysis, latched as a group: one entry per configured slot, in register order.
const note = ref(AUTO_NOTE)
const noteName = ref('')
const scalePitch = ref(0)
const label = ref('')
const reeds = shallowRef<ReedMeasure[]>([])
const beats = shallowRef<BeatMeasure[]>([])

// The sounding slots scored against the session's goal curve; no session = plain scale deviation.
const reedErrors = shallowRef<ReedError[]>([])
const beatErrors = shallowRef<BeatError[]>([])

// The reference screen state, on every frame: readouts reed, pair, reed..., the retained needles and the lock latches.
const panels = shallowRef<Panel[]>([])
const needles = shallowRef<Needle[]>([])
const needlesDimmed = ref(true)
const lock = shallowRef<Lock>(NO_LOCK)

const inputLevel = ref(0)
const unreliable = ref(false)
const blink = ref(false)
const state = ref<EngineState>('stopped')
const running = ref(false)

// Where in the recording the on-screen reading was measured, or null when there is no answer.
const readingAt = ref<number | null>(null)
const starting = ref(false)
const error = ref('')

// What the engine measures with, kept live by EVENT_SETTINGS: never derived here.
const settings = tunerSettings()
const { a4, manualNote, transpose } = settings
const frozen = ref(false)

// Freeze needs a reading to hold: while running, or after a stop left its last note.
const canFreeze = computed(() => running.value || note.value !== AUTO_NOTE)

const tunerState: TunerState = {
  note, noteName, scalePitch, label, reeds, beats, reedErrors, beatErrors, panels, needles, needlesDimmed, lock,
  inputLevel, unreliable, blink, state, running, readingAt, starting, error,
  live,
  notify: () => { for (const fn of listeners) fn() },
}

// Wipe the screen. Called at every change of source: see clearReading.
const clearReading = () => clearReadingImpl(tunerState)

const scope = effectScope(true)
let started = false

type Logger = ReturnType<typeof useLogger>

function start(log: Logger) {
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
      applySettings(ev.data, settings)
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
  })

  // Catch settings the engine resolved before this listener was up.
  TunerService.Settings()
    .then(s => applySettings(s, settings))
    .catch(() => {})

  TunerService.IsRunning()
    .then((on) => {
      running.value = on
      if (on) state.value = 'running'
    })
    .catch((err) => {
      log.error('log.frontendTunerStateFailed', { error: String(err) })
    })
}

// pin shows the key the service pinned at once; tuner:settings repeats it.
async function pin(call: () => Promise<number>, log: Logger, what: string) {
  const n = await command(tunerState, call, log, what)
  if (n !== undefined) manualNote.value = n
}

export function useTuner() {
  const log = useLogger()
  start(log)

  return {
    note,
    noteName,
    scalePitch,
    label,
    reeds,
    beats,
    reedErrors,
    beatErrors,
    panels,
    needles,
    needlesDimmed,
    lock,
    inputLevel,
    unreliable,
    blink,
    state,

    running,
    starting,
    readingAt,
    clearReading,
    error,
    start: () => startEngine(tunerState, log),
    stop: () => stopEngine(tunerState, log),

    ...settings,
    frozen,
    canFreeze,
    setA4: (hz: number) => apply(tunerState, a4, hz, () => TunerService.SetA4(hz), log, 'a4'),
    setManualNote: (n: number) => apply(tunerState, manualNote, n, () => TunerService.SetManualNote(n), log, 'manual note'),
    setTranspose: (semitones: number) => apply(tunerState, transpose, semitones, () => TunerService.SetTranspose(semitones), log, 'transpose'),
    setFrozen: (on: boolean) => apply(tunerState, frozen, on, () => TunerService.Freeze(on), log, 'freeze'),
    // The service picks the note (tuning.ManualStart, Note.Step); the answer is the pinned key.
    setAuto: (on: boolean) => pin(() => TunerService.SetAuto(on), log, 'auto'),
    stepNote: (delta: number) => pin(() => TunerService.StepNote(delta), log, 'note step'),
    setMode: (mode: PanelMode) => command(tunerState, () => TunerService.SetMode(mode), log, 'mode'),
    setReedBox: (box: number, on: boolean) => command(tunerState, () => TunerService.SetReedBox(box, on), log, 'reed box'),
    setOctave: (offset: number, on: boolean) => command(tunerState, () => TunerService.SetOctave(offset, on), log, 'octave box'),
    takeNow: () => command(tunerState, () => TunerService.TakeNow(), log, 'take now'),
    relearn: () => command(tunerState, () => TunerService.Recalibrate(), log, 'relearn'),
    playTone: (n: number) => playTone(tunerState, n, log),
    stopTone: () => stopTone(log),
  }
}
