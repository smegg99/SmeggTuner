import { beforeEach, describe, expect, it, vi } from 'vitest'

/*
 * CHANGING THE SOURCE WIPES THE SCREEN.
 *
 * Every reading on it was measured from the input that has just been replaced. The reeds,
 * the note, the deviation, the spectrum, the note strip, the input trace and the ghost mark
 * all describe audio that is no longer the input - and not one of them LOOKS stale. They
 * look like a reading of the recording you are now staring at.
 */

const clearReading = vi.fn()
const refresh = vi.fn(async (_sourceChanged?: boolean) => {})

vi.mock('~/composables/useTuner', () => ({
  useTuner: () => ({ clearReading }),
  toErrorKey: () => 'tuner.error.unexpected',
}))

vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({ refresh }),
}))

const SelectMic = vi.fn(async () => {})
const SelectFile = vi.fn(async () => {})

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js', () => ({
  ListDevices: vi.fn(async () => []),
  Current: vi.fn(async () => ({ kind: 'mic', deviceId: '', path: '', loop: false, name: '' })),
  SelectMic,
  SelectFile,
  OpenFileDialog: vi.fn(async () => ''),
}))

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js', () => ({
  IsRunning: vi.fn(async () => false),
  Restart: vi.fn(async () => {}),
}))

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js', () => ({
  SourceKind: { SourceFile: 'file', SourceMic: 'mic' },
}))

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useLogger = () => ({ error: vi.fn(), info: vi.fn(), debug: vi.fn() })

// The native file picker is handed its words, because Go has none: see
// services/audio.OpenFileDialog. Nothing here is about the wording, so the key will do.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t: (key: string) => key })

const { useAudioDevices } = await import('~/composables/useAudioDevices')

describe('changing the input', () => {
  beforeEach(() => {
    clearReading.mockClear()
  })

  it('wipes the reading when a recording is loaded', async () => {
    await useAudioDevices().selectFile('/tmp/other.wav', false)
    expect(clearReading).toHaveBeenCalled()
    expect(refresh).toHaveBeenLastCalledWith(true)
  })

  it('wipes the reading when the microphone is chosen', async () => {
    await useAudioDevices().selectMic('some-device')
    expect(clearReading).toHaveBeenCalled()
  })

  // The reading is wiped BEFORE the new source is drawn, not after: a panel that is briefly
  // showing the old file's reeds over the new file's waveform is the exact frame this
  // exists to prevent.
  it('wipes it before the new source is adopted', async () => {
    const order: string[] = []
    clearReading.mockImplementation(() => order.push('cleared'))
    refresh.mockImplementation(async () => {
      order.push('adopted')
    })

    await useAudioDevices().selectFile('/tmp/other.wav', false)

    expect(order).toEqual(['cleared', 'adopted'])
  })
})
