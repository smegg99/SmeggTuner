// The selected capture device's sound-card clock calibration and automatic volume (services/audio).
import { computed, effectScope, ref, shallowRef, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import * as AudioService from '~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js'
import type { AutoVolumeDTO, CalibrationDTO, CorrectionDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useConfigSync } from '~/composables/useConfigSync'
import { toErrorKey } from '~/composables/useTuner'

const EVENT_CALIBRATION = 'audio:calibration'
const EVENT_VOLUME = 'audio:volume'

const calibration = shallowRef<CalibrationDTO | null>(null)
const autoVolume = shallowRef<AutoVolumeDTO | null>(null)
// The current input's correction: what services/tuner runs with.
const correction = shallowRef<CorrectionDTO | null>(null)
// The shown device's clock offset over the current or last run, one point per update.
export interface TracePoint { at: number, percent: number }
const trace = shallowRef<TracePoint[]>([])
const TRACE_MAX = 1200
const commandError = ref('')
const error = computed(() => calibration.value?.error || commandError.value)

// Devices asked about at mic start in this run: the question comes once per device.
const asked = new Set<string>()

const scope = effectScope(true)
let started = false
let seq = 0
let requestedDevice: string | undefined
let pendingCalibrationEvents: Map<string, CalibrationDTO> | null = null

type Logger = ReturnType<typeof useLogger>

async function refreshCorrection(log: Logger) {
  try {
    correction.value = await AudioService.Correction()
  }
  catch (err) {
    log.error('log.frontendAudioCorrectionFailed', { error: String(err) })
  }
}

// refresh reads the device's state; only the newest read may write, the selection can move under it.
async function refresh(deviceId: string, log: Logger) {
  const mine = ++seq
  const events = new Map<string, CalibrationDTO>()
  pendingCalibrationEvents = events
  if (requestedDevice !== deviceId) {
    requestedDevice = deviceId
    calibration.value = null
    trace.value = []
    autoVolume.value = null
    commandError.value = ''
  }
  try {
    const [cal, vol] = await Promise.all([AudioService.CalibrationState(deviceId), AudioService.AutoVolume(deviceId)])
    if (mine !== seq) return
    // The read resolves the system-default alias; retain any newer event for that concrete device.
    calibration.value = events.get(cal.deviceId) ?? cal
    autoVolume.value = vol
  }
  catch (err) {
    if (mine === seq) log.error('log.frontendAudioStateFailed', { error: String(err) })
  }
  finally {
    if (mine === seq) pendingCalibrationEvents = null
  }
}

// A run starts warming up, so that clears the trace; only measured blocks carry an offset.
function record(c: CalibrationDTO) {
  if (!c.running) return
  if (c.status === 'preparing') trace.value = []
  else if (c.status === 'calibrating') trace.value = [...trace.value, { at: c.elapsed, percent: c.percent }].slice(-TRACE_MAX)
}

export function useSoundCard() {
  const log = useLogger()
  const { current } = useAudioDevices()
  const { config } = useConfigSync()

  // The device the audio settings show: the selected mic, or the saved one while a recording is the input.
  const deviceId = computed(() => (current.value.kind === 'mic' ? current.value.deviceId : config.audio?.device_id ?? ''))

  if (!started) {
    started = true
    scope.run(() => {
      Events.On(EVENT_CALIBRATION, (ev: { data: CalibrationDTO }) => {
        pendingCalibrationEvents?.set(ev.data.deviceId, ev.data)
        const was = calibration.value
        if (ev.data.deviceId === was?.deviceId || ev.data.deviceId === deviceId.value) {
          calibration.value = ev.data
          record(ev.data)
        }
        // A session stores the missing marker first and the outcome last; the input's correction follows both.
        if (!was || was.status !== ev.data.status || !ev.data.running) void refreshCorrection(log)
      })
      Events.On(EVENT_VOLUME, (ev: { data: AutoVolumeDTO }) => {
        if (autoVolume.value?.deviceId === ev.data.deviceId) autoVolume.value = ev.data
      })
      watch(deviceId, id => void refresh(id, log), { immediate: true, flush: 'sync' })
      watch(() => [current.value.kind, current.value.deviceId], () => void refreshCorrection(log), { immediate: true })
    })
  }

  async function startCalibration() {
    commandError.value = ''
    try {
      await AudioService.StartCalibration(deviceId.value)
    }
    catch (err) {
      commandError.value = /settings\.calibration\.[A-Za-z0-9]+/.exec(String(err))?.[0] ?? toErrorKey(err)
      log.error('log.frontendAudioCalibrationFailed', { error: String(err) })
    }
    await refresh(deviceId.value, log)
  }

  async function cancelCalibration() {
    await AudioService.CancelCalibration().catch(err => log.error('log.frontendAudioCalibrationCancelFailed', { error: String(err) }))
    await refresh(deviceId.value, log)
  }

  async function setAutoVolume(on: boolean) {
    try {
      autoVolume.value = await AudioService.SetAutoVolume(deviceId.value, on)
    }
    catch (err) {
      log.error('log.frontendAudioVolumeFailed', { error: String(err) })
    }
  }

  // needsAsking: a mic start on a device never calibrated asks once whether to measure it first.
  async function needsAsking(): Promise<boolean> {
    if (current.value.kind !== 'mic') return false
    try {
      const state = await AudioService.CalibrationState(current.value.deviceId)
      if (asked.has(state.deviceId) || state.status !== 'uncalibrated') return false
      asked.add(state.deviceId)
      return true
    }
    catch {
      return false
    }
  }

  return {
    deviceId,
    calibration,
    trace,
    autoVolume,
    correction,
    error,
    // The mic in use was never calibrated: it measures uncorrected, so the app asks for a calibration.
    uncalibrated: computed(() => correction.value?.missing === true),
    // A recording is never steered, so its clipping is always worth saying.
    steering: computed(() => current.value.kind === 'mic' && autoVolume.value?.on === true),
    startCalibration,
    cancelCalibration,
    setAutoVolume,
    needsAsking,
  }
}
