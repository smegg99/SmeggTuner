import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { effectScope, ref } from 'vue'
import FileWave from '~/components/file/FileWave.vue'
import { useTransport } from '~/composables/useTransport'
import { useAudioDevices } from '~/composables/useAudioDevices'
import type { PeakDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'

const backend = vi.hoisted(() => ({ peaks: vi.fn(), select: vi.fn() }))
const dto = {
  available: true, name: 'note.wav', duration: 10, position: 0,
  from: 0, to: 10, paused: true, moving: false, loop: false, sampleRate: 48000,
}

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js', () => ({
  Transport: async () => ({ ...dto }), Muted: async () => false, Volume: async () => 1,
  Peaks: backend.peaks, SelectFile: backend.select, ListDevices: async () => [],
  Current: async () => ({ kind: 'file', deviceId: '', path: '/two/note.wav', name: 'note.wav', loop: false }),
}))
vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js', () => ({ IsRunning: async () => false }))
vi.mock('~/composables/useTuner', () => ({
  useTuner: () => ({ running: ref(false), readingAt: ref(null), clearReading: vi.fn() }),
  toErrorKey: () => 'error',
}))
vi.mock('~/composables/useCanvas', () => ({
  useCanvas: () => ({ host: ref(null), canvas: ref(null), paint: vi.fn(), redraw: vi.fn() }),
}))
vi.mock('~/composables/render/useWaveRender', () => ({ useWaveRender: () => ({}), laneHeight: () => 20 }))
vi.mock('vuetify', () => ({ useTheme: () => ({ global: { name: ref('light') } }) }))

const oldPeaks = [{ min: -0.1, max: 0.1 }] as PeakDTO[]
const newPeaks = [{ min: -0.8, max: 0.8 }] as PeakDTO[]
let scope: ReturnType<typeof effectScope>
let transport: ReturnType<typeof useTransport>
let wrapper: ReturnType<typeof mount> | undefined

beforeEach(async () => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  vi.stubGlobal('useLogger', () => ({ error: vi.fn() }))
  vi.stubGlobal('requestAnimationFrame', () => 1)
  vi.stubGlobal('cancelAnimationFrame', vi.fn())
  vi.spyOn(HTMLElement.prototype, 'clientWidth', 'get').mockReturnValue(100)
  backend.peaks.mockReset().mockResolvedValue(oldPeaks)
  backend.select.mockReset().mockResolvedValue(undefined)
  scope = effectScope()
  transport = scope.run(() => useTransport())!
  await transport.refresh(true)
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  scope.stop()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('waveform source and request identity', () => {
  it('reloads the mounted waveform after successful identical-metadata source replacement', async () => {
    const devices = scope.run(() => useAudioDevices())!
    await flushPromises()
    wrapper = mount(FileWave)
    await flushPromises()
    expect(transport.peaks.value).toEqual(oldPeaks)
    const calls = backend.peaks.mock.calls.length

    backend.peaks.mockResolvedValue(newPeaks)
    await devices.selectFile('/two/note.wav', false)
    await flushPromises()
    expect(transport.state).toMatchObject(dto)
    expect(backend.peaks.mock.calls.length).toBeGreaterThan(calls)
    expect(transport.peaks.value).toEqual(newPeaks)

    // Reopening the same path also replaces samples; path comparison is insufficient.
    backend.peaks.mockResolvedValue(oldPeaks)
    await devices.selectFile('/two/note.wav', false)
    await flushPromises()
    expect(transport.peaks.value).toEqual(oldPeaks)
  })

  it('rejects a previous source reply even before the new waveform request starts', async () => {
    let resolve!: (peaks: PeakDTO[]) => void
    backend.peaks.mockReturnValueOnce(new Promise<PeakDTO[]>((done) => {
      resolve = done
    }))
    const old = transport.loadPeaks(100)
    await transport.refresh(true)
    resolve(oldPeaks)
    await old
    expect(transport.peaks.value).toEqual([])
    backend.peaks.mockResolvedValue(newPeaks)
    await transport.loadPeaks(100)
    expect(transport.peaks.value).toEqual(newPeaks)
  })

  it('keeps the latest range/width reply when requests finish out of order', async () => {
    let resolve!: (peaks: PeakDTO[]) => void
    backend.peaks.mockReturnValueOnce(new Promise<PeakDTO[]>((done) => {
      resolve = done
    }))
    const old = transport.loadPeaks(100)
    transport.zoom(0.5, 5)
    backend.peaks.mockResolvedValue(newPeaks)
    await transport.loadPeaks(200)
    resolve(oldPeaks)
    await old
    expect(transport.peaks.value).toEqual(newPeaks)
  })

  it('rejects a reply after the view changes before its next reload', async () => {
    let resolve!: (peaks: PeakDTO[]) => void
    backend.peaks.mockReturnValueOnce(new Promise<PeakDTO[]>((done) => {
      resolve = done
    }))
    const old = transport.loadPeaks(100)
    transport.zoom(0.5, 5)
    resolve(oldPeaks)
    await old
    expect(transport.peaks.value).toEqual([])
  })
})
