// tests/web/storedDevice.spec.ts
import { describe, expect, it } from 'vitest'
import { storedDeviceId, storedPreferences } from '~/web/storage'

describe('storedDeviceId', () => {
  it('reads the saved microphone from the stored config', () => {
    expect(storedDeviceId('{"audio":{"device_id":"usb"}}')).toBe('usb')
  })

  it('is the default for empty or broken config', () => {
    expect(storedDeviceId('')).toBe('')
    expect(storedDeviceId('{')).toBe('')
    expect(storedDeviceId('{"audio":{"device_id":7}}')).toBe('')
  })
})

describe('storedPreferences', () => {
  it('reads theme and language', () => {
    expect(storedPreferences('{"preferences":{"theme":"dark","language":"pl"}}')).toEqual({ theme: 'dark', language: 'pl' })
  })

  it('falls back to auto and empty language', () => {
    expect(storedPreferences('')).toEqual({ theme: 'auto', language: '' })
    expect(storedPreferences('{"preferences":{"theme":5}}')).toEqual({ theme: 'auto', language: '' })
  })
})
