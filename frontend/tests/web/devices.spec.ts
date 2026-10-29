// tests/web/devices.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('~/utils/features', () => ({ WEB: true }))
vi.mock('~/composables/useTuner', () => ({
  useTuner: () => ({ clearReading: vi.fn() }),
  toErrorKey: () => 'tuner.error.unexpected',
}))
vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({ refresh: async () => {} }),
}))

const { list, current } = vi.hoisted(() => ({
  list: vi.fn(async () => [{ id: 'usb', name: 'USB microphone', default: false }]),
  current: vi.fn(async () => ({ kind: 'mic', deviceId: '', path: '', loop: false, name: '' })),
}))
vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js', () => ({
  ListDevices: list,
  Current: current,
}))

beforeEach(() => {
  vi.resetModules()
  list.mockClear()
  current.mockClear()
  vi.stubGlobal('useLogger', () => ({ error: vi.fn() }))
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
})
afterEach(() => vi.unstubAllGlobals())

async function devices() {
  const { useAudioDevices } = await import('~/composables/useAudioDevices')
  const audio = useAudioDevices()
  await vi.waitFor(() => expect(audio.loading.value).toBe(false))
  return audio
}

function media() {
  const stop = vi.fn()
  const stream = { getTracks: () => [{ stop }] } as unknown as MediaStream
  const getUserMedia = vi.fn(async () => stream)
  vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
  return { stop, stream, getUserMedia }
}

describe('browsing browser input devices', () => {
  it('does not request permission on mount or a background refresh', async () => {
    const { getUserMedia } = media()
    const audio = await devices()
    await audio.refresh()
    expect(getUserMedia).not.toHaveBeenCalled()
  })

  it('requests in the click call, waits for permission, then lists and releases the input', async () => {
    const { getUserMedia, stream, stop } = media()
    const audio = await devices()
    list.mockClear()
    let grant!: (stream: MediaStream) => void
    getUserMedia.mockImplementationOnce(() => new Promise((resolve) => {
      grant = resolve
    }))

    const pending = audio.browse()
    expect(getUserMedia).toHaveBeenCalledTimes(1)
    expect(audio.loading.value).toBe(true)
    expect(list).not.toHaveBeenCalled()
    await audio.browse()
    expect(getUserMedia).toHaveBeenCalledTimes(1)

    list.mockImplementationOnce(async () => {
      expect(stop).not.toHaveBeenCalled()
      return [{ id: 'usb', name: 'USB microphone', default: false }]
    })
    grant(stream)
    await pending
    expect(audio.devices.value[0]?.name).toBe('USB microphone')
    expect(stop).toHaveBeenCalledTimes(1)
    expect(audio.loading.value).toBe(false)
  })

  it('shows a localized permission error and allows retry', async () => {
    const { getUserMedia, stop } = media()
    const audio = await devices()
    list.mockClear()
    getUserMedia.mockRejectedValueOnce(new DOMException('Denied', 'NotAllowedError'))
    await audio.browse()
    expect(audio.error.value).toBe('web.mic.denied')
    expect(audio.loading.value).toBe(false)
    expect(list).not.toHaveBeenCalled()
    expect(stop).not.toHaveBeenCalled()

    await audio.browse()
    expect(audio.error.value).toBe('')
    expect(stop).toHaveBeenCalledTimes(1)
  })

  it('releases the temporary microphone when enumeration fails', async () => {
    const { stop } = media()
    const audio = await devices()
    list.mockRejectedValueOnce(new Error('Device enumeration failed'))
    await audio.browse()
    expect(stop).toHaveBeenCalledTimes(1)
    expect(audio.error.value).toBe('tuner.error.unexpected')
    expect(audio.loading.value).toBe(false)
  })
})
