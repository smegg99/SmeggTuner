import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

/*
 * THE GHOST, AND EVERY WAY IT COULD LIE.
 *
 * It marks where the reading standing on screen was measured. That is a claim about the
 * past, made on a picture of a recording, and there are exactly three ways to make it
 * false. All three are refused in one place - useTransport.ghostAt - and all three are
 * here.
 */

const running = ref(false)
const readingAt = ref<number | null>(null)
const clearReading = vi.fn(() => {
  readingAt.value = null
})

vi.mock('~/composables/useTuner', () => ({
  useTuner: () => ({ running, readingAt, clearReading }),
  toErrorKey: () => '',
}))

// A real DTO, not undefined: useTransport re-asks Go for the transport whenever the engine
// stops, and a mock that answers with nothing throws inside a watcher - which vitest reports
// as an unhandled error while still passing every assertion.
const DTO = {
  available: true,
  name: 'h-16.wav',
  duration: 9,
  position: 4,
  from: 0,
  to: 9,
  paused: true,
  moving: false,
  loop: false,
  sampleRate: 48000,
}

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js', () => ({
  Transport: vi.fn(async () => DTO),
  Muted: vi.fn(async () => false),
  Volume: vi.fn(async () => 1),
  Seek: vi.fn(),
  SetPaused: vi.fn(),
  SetLoop: vi.fn(),
  SetRange: vi.fn(),
  SetVolume: vi.fn(),
  SetMuted: vi.fn(),
  Peaks: vi.fn(),
}))

vi.mock('@wailsio/runtime', () => ({ Events: { On: vi.fn() } }))

const { useTransport } = await import('~/composables/useTransport')

describe('the ghost', () => {
  let t: ReturnType<typeof useTransport>

  beforeEach(() => {
    t = useTransport()

    Object.assign(t.state, {
      available: true,
      duration: 9,
      to: 9,
      paused: true,
      moving: false,
    })
    running.value = false
    readingAt.value = 4.2
    t.dragging.value = false
  })

  it('marks exactly where the reading was measured', () => {
    expect(t.ghostAt.value).toBe(4.2)
  })

  /*
   * IT DOES NOT SHOW WHILE IT IS PLAYING. The numbers are live and are being replaced four
   * times a second; a mark chasing them about the waveform explains nothing and is simply
   * one more thing moving. It appears when the sound stops - which is the only moment the
   * question is asked, and the only moment the answer stands still.
   */
  it('vanishes the moment the audio is moving again', () => {
    running.value = true
    t.state.paused = false
    t.state.moving = true

    expect(t.ghostAt.value).toBeNull()
  })

  it('comes back when the sound stops', () => {
    running.value = true
    t.state.paused = false
    t.state.moving = true
    expect(t.ghostAt.value).toBeNull()

    t.state.paused = true // paused: the sound has stopped and the reading has stood still
    t.state.moving = false
    expect(t.ghostAt.value).toBe(4.2)
  })

  /*
   * A MICROPHONE HAS NO TIMELINE. Go reports zero for a room, and the frontend keeps that
   * as null rather than as a number - a mark drawn on `> 0` is a mark that eventually gets
   * drawn on a reading whose true answer was "there isn't one".
   */
  it('points nowhere when there is no reading to point at', () => {
    readingAt.value = null
    expect(t.ghostAt.value).toBeNull()
  })

  it('points nowhere when there is no recording', () => {
    t.state.available = false
    expect(t.ghostAt.value).toBeNull()
  })

  /*
   * AND IT NEVER OUTLIVES THE RECORDING IT CAME FROM.
   *
   * A position measured in one file, drawn on the waveform of another, is the worst kind of
   * wrong: precise, plausible, and fiction. Nothing about it looks wrong. useAudioDevices
   * drops it on every source change, and this is the guarantee that makes that safe.
   */
  it('is dropped when the recording changes', () => {
    expect(t.ghostAt.value).toBe(4.2)

    clearReading() // what useAudioDevices.select does on every source change
    expect(t.ghostAt.value).toBeNull()
  })
})
