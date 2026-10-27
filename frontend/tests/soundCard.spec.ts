import { describe, expect, it } from 'vitest'
import type { CalibrationDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'
import { calibrationLine, calibrationTone, clockPitch, signedPercent, tooStrongShown } from '~/utils/soundCard'

const dto = (o: Partial<CalibrationDTO>): CalibrationDTO =>
  ({ deviceId: 'a', status: 'uncalibrated', percent: 0, elapsed: 0, progress: 0, running: false, ...o }) as CalibrationDTO

describe('the sound-card calibration display', () => {
  it('colours the states as the reference does', () => {
    expect(calibrationTone('calibrated')).toBe('goal')
    expect(calibrationTone('preparing')).toBe('accent')
    expect(calibrationTone('calibrating')).toBe('accent')
    expect(calibrationTone('uncalibrated')).toBe('warn')
    expect(calibrationTone('unreliable')).toBe('reed')
  })

  it('prints a correction to three decimals with a plus when positive', () => {
    expect(signedPercent(0.0123)).toBe('+0.012')
    expect(signedPercent(-0.5)).toBe('-0.500')
    expect(signedPercent(0)).toBe('0.000')
  })

  it('says how long a session has run and how far off the clock is', () => {
    const line = calibrationLine(dto({ status: 'calibrating', elapsed: 23, percent: 0.0121, progress: 61, running: true }))
    expect(line.key).toBe('settings.calibration.status.calibrating')
    expect(line.params).toEqual({ seconds: 23, percent: '+0.012', progress: 61 })
  })

  // 1 reference unit is 0.06 % of the frequency; a clock that far off reads every note 1 u (1.04 cent) sharp.
  it('restates a clock correction as the pitch error it stands for', () => {
    expect(clockPitch(0.06, 'reference')).toBe('+1.00')
    expect(clockPitch(0.06, 'cents')).toBe('+1.04')
    expect(clockPitch(-0.003, 'reference')).toBe('-0.05')
  })

  it('reads an unknown status as uncalibrated', () => {
    expect(calibrationLine(dto({ status: 'odd' as CalibrationDTO['status'] })).key).toBe('settings.calibration.status.uncalibrated')
  })
})

// "Input too strong": a flat-top interval, shown on alternate analyses, never while automatic volume steers.
describe('the clip warning', () => {
  it('blinks only with automatic volume off', () => {
    expect(tooStrongShown({ state: 'tooLoud', autoVolume: false, blink: true })).toBe(true)
    expect(tooStrongShown({ state: 'tooLoud', autoVolume: false, blink: false })).toBe(false)
    expect(tooStrongShown({ state: 'tooLoud', autoVolume: true, blink: true })).toBe(false)
    expect(tooStrongShown({ state: 'running', autoVolume: false, blink: true })).toBe(false)
  })
})
