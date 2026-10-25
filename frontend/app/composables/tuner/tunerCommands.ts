import type { Ref } from 'vue'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js'
import { toErrorKey } from './tunerProtocol'
import { clearReading } from './tunerReducers'
import type { TunerState } from './tunerReducers'

type Logger = ReturnType<typeof useLogger>

// apply moves the ref optimistically and rolls back on rejection unless something newer already moved it.
export async function apply<T>(s: TunerState, target: Ref<T>, next: T, call: () => Promise<void>, log: Logger, what: string) {
  const previous = target.value
  target.value = next

  try {
    await call()
  }
  catch (err) {
    if (target.value === next) target.value = previous
    s.error.value = toErrorKey(err)
    log.error('log.frontendSettingFailed', { setting: what, error: String(err) })
  }
}

// command runs a setter whose outcome arrives as tuner:settings; a refusal lands in error.
export async function command<T>(s: TunerState, call: () => Promise<T>, log: Logger, what: string): Promise<T | undefined> {
  try {
    return await call()
  }
  catch (err) {
    s.error.value = toErrorKey(err)
    log.error('log.frontendSettingFailed', { setting: what, error: String(err) })
    return undefined
  }
}

export async function startEngine(s: TunerState, log: Logger) {
  if (s.running.value || s.starting.value) return

  s.starting.value = true
  clearReading(s)
  s.error.value = ''
  // Learning is the engine's to report: a file, or a microphone that already learned its room, never does.
  s.state.value = 'running'

  try {
    await TunerService.Start()
    s.running.value = true
  }
  catch (err) {
    s.running.value = false
    s.state.value = 'stopped'
    s.error.value = toErrorKey(err)
    log.error('log.frontendTunerStartFailed', { error: String(err) })
  }
  finally {
    s.starting.value = false
  }
}

export async function stopEngine(s: TunerState, log: Logger) {
  try {
    await TunerService.Stop()
  }
  catch (err) {
    s.error.value = toErrorKey(err)
    log.error('log.frontendTunerStopFailed', { error: String(err) })
    return
  }

  // The last picture stays up after stop (see onState).
  s.running.value = false
  s.state.value = 'stopped'
  s.inputLevel.value = 0
}

export async function playTone(s: TunerState, toneNote: number, log: Logger) {
  try {
    await TunerService.PlayTone(toneNote)
  }
  catch (err) {
    s.error.value = toErrorKey(err)
    log.error('log.frontendTunerTonePlayFailed', { error: String(err) })
  }
}

// stopTone ends the hold-to-play reference tone; idempotent and never surfaces an error.
export async function stopTone(log: Logger) {
  try {
    await TunerService.StopTone()
  }
  catch (err) {
    log.error('log.frontendTunerToneStopFailed', { error: String(err) })
  }
}
