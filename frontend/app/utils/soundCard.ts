import type { CalibrationDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'
import { toCents } from '~/utils/deviation'
import type { DeviationUnit } from '~/utils/deviation'

// Display rules for the sound-card calibration and the input warnings; the numbers are services/audio's.

export type CalibrationTone = 'goal' | 'accent' | 'warn' | 'reed'

/** The reference tuner's status colours: green calibrated, blue running, orange uncalibrated, red unreliable. */
export function calibrationTone(status: string): CalibrationTone {
  switch (status) {
    case 'calibrated':
      return 'goal'
    case 'preparing':
    case 'calibrating':
      return 'accent'
    case 'unreliable':
      return 'reed'
  }
  return 'warn'
}

/** A clock correction as the reference prints it: three decimals, a plus when positive. */
export function signedPercent(p: number): string {
  const text = p.toFixed(3)
  return p > 0 ? `+${text}` : text
}

/** clockPitch: a clock correction in percent as the pitch error it stands for at every note, signed, two decimals. */
export function clockPitch(percent: number, unit: DeviationUnit): string {
  const units = percent / 0.06
  const v = unit === 'cents' ? toCents(units) : units
  const text = v.toFixed(2)
  return v > 0 ? `+${text}` : text
}

export interface CalibrationLine {
  /** settings.calibration.status.* */
  key: string
  params: Record<string, string | number>
}

/** calibrationLine is the status sentence for a device: running sessions say how long and how far off. */
export function calibrationLine(c: CalibrationDTO): CalibrationLine {
  const status = ['uncalibrated', 'preparing', 'calibrating', 'calibrated', 'unreliable'].includes(c.status) ? c.status : 'uncalibrated'
  return {
    key: `settings.calibration.status.${status}`,
    params: { seconds: c.elapsed, percent: signedPercent(c.percent), progress: c.progress },
  }
}

export interface InputWarnings {
  /** the engine's state: tooLoud is the flat-top latch of the last interval */
  state: string
  /** automatic volume steers this input; the reference never shows the warning then */
  autoVolume: boolean
  /** flips on every analysis */
  blink: boolean
}

/** tooStrongShown: the clip warning, only with automatic volume off, on alternate analyses. */
export function tooStrongShown(w: InputWarnings): boolean {
  return w.state === 'tooLoud' && !w.autoVolume && w.blink
}
