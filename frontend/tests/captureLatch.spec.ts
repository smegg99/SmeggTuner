// @vitest-environment node
// tests/captureLatch.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref } from 'vue'

const note = ref(0)
const lock = ref({ visible: false })

vi.mock('~/composables/useTuner', () => ({ useTuner: () => ({ note, lock }) }))
vi.mock('~/composables/useConfigSync', () => ({
  useConfigSync: () => ({ config: { tuner: { calibration_capture_ms: 2000 } } }),
}))

const { useCaptureLatch } = await import('~/composables/captureLatch')
const latch = useCaptureLatch()

async function hear(n: number) {
  note.value = n
  lock.value = { visible: true }
  await nextTick()
}

async function release() {
  lock.value = { visible: false }
  await nextTick()
}

beforeEach(async () => {
  vi.useFakeTimers()
  await release()
  note.value = 0
  latch.resetLatch()
})

afterEach(() => {
  latch.resetLatch()
  vi.useRealTimers()
})

describe('calibration capture latch', () => {
  it('allows the next note after consuming a released note during grace', async () => {
    await hear(48)
    await release()
    expect(latch.heard.value).toBe(48)
    latch.consume()
    expect(latch.heard.value).toBe(0)

    await hear(84)
    expect(latch.heard.value).toBe(84)
    await release()
    expect(latch.heard.value).toBe(84)
  })

  it('suppresses the consumed live lock and its release tail, then accepts a new lock', async () => {
    await hear(48)
    latch.consume()
    expect(latch.heard.value).toBe(0)
    await hear(48)
    expect(latch.heard.value).toBe(0)
    await release()
    expect(latch.heard.value).toBe(0)
    await vi.advanceTimersByTimeAsync(2000)
    expect(latch.heard.value).toBe(0)

    await hear(84)
    expect(latch.heard.value).toBe(84)
  })

  it('expires an unconsumed grace note after the configured interval', async () => {
    await hear(48)
    await release()
    await vi.advanceTimersByTimeAsync(1999)
    expect(latch.heard.value).toBe(48)
    await vi.advanceTimersByTimeAsync(1)
    expect(latch.heard.value).toBe(0)
    expect(latch.shown.value).toBe(48)
  })

  it('clears the grace note and its timer when the range is reset', async () => {
    await hear(48)
    await release()
    latch.resetLatch()
    expect(latch.heard.value).toBe(0)
    expect(latch.shown.value).toBe(0)
    expect(vi.getTimerCount()).toBe(0)
  })
})
