// tests/web/local.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { bindingID, localServices } from '~/web/local'

function mic(deviceId = '') {
  return { deviceId, select: vi.fn(async () => {}), label: 'Built-in' } as unknown as Parameters<typeof localServices>[0]
}

afterEach(() => vi.unstubAllGlobals())

describe('browser-side services', () => {
  it('computes the same binding IDs as the generator', () => {
    // From bindings/smegg.me/smeggtuner/backend/services/audio/service.ts.
    expect(bindingID('smegg.me/smeggtuner/backend/services/audio.Service.ListDevices')).toBe(1670568525)
    expect(bindingID('smegg.me/smeggtuner/backend/services/audio.Service.SelectMic')).toBe(169356577)
  })

  it('lists browser microphones as DeviceDTOs, without the browser\'s default alias', async () => {
    vi.stubGlobal('navigator', { ...navigator, mediaDevices: { enumerateDevices: async () => [
      { kind: 'audioinput', deviceId: 'default', label: 'Default' },
      { kind: 'audioinput', deviceId: 'usb', label: 'USB mic' },
      { kind: 'videoinput', deviceId: 'cam', label: 'Camera' },
    ] } })
    const list = await localServices(mic()).get(1670568525)!([])
    // The browser's own "default" entry duplicates the app's System default, which is the empty id.
    expect(list).toEqual([{ id: 'usb', name: 'USB mic', default: false }])
  })

  it('selecting a microphone goes to the browser input and is saved', async () => {
    const m = mic()
    const save = vi.fn(async () => {})
    await localServices(m, save).get(169356577)!(['usb'])
    expect(m.select).toHaveBeenCalledWith('usb')
    expect(save).toHaveBeenCalledWith('usb')
  })

  it('never asks for a sound-card calibration', async () => {
    const state = await localServices(mic('usb')).get(2151603095)!(['usb']) as { status: string }
    expect(state.status).not.toBe('uncalibrated')
  })
})

describe('browser microphone names', () => {
  it('names the current microphone from the device list before it is opened', async () => {
    vi.stubGlobal('navigator', { ...navigator, mediaDevices: { enumerateDevices: async () => [
      { kind: 'audioinput', deviceId: 'usb', label: 'USB mic' },
    ] } })
    const services = localServices({ deviceId: 'usb', label: '', select: async () => {} })
    await services.get(1670568525)!([])
    const current = await services.get(bindingID('smegg.me/smeggtuner/backend/services/audio.Service.Current'))!([]) as { name: string }
    expect(current.name).toBe('USB mic')
  })
})

describe('reference tone in the browser', () => {
  it('is a silent no-op, not an error', async () => {
    const services = localServices({ deviceId: '', label: '', select: async () => {} })
    for (const m of ['PlayTone', 'StopTone', 'ToggleTone']) {
      await expect(Promise.resolve(services.get(bindingID(`smegg.me/smeggtuner/backend/services/tuner.Service.${m}`))!([69]))).resolves.toBeNull()
    }
  })
})

describe('theme in the browser', () => {
  it('reports no system accent, since a browser cannot read one', async () => {
    const info = await localServices({ deviceId: '', label: '', select: async () => {} })
      .get(bindingID('github.com/smegg99/s99wails/v2/services/theme.Service.GetTheme'))!([]) as { accent_color: string, accent_watch_supported: boolean }
    expect(info.accent_color).toBe('')
    expect(info.accent_watch_supported).toBe(false)
  })
})
