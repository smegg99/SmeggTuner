import { afterEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref } from 'vue'
import { useReadoutStatus, type ReadoutStatus } from '~/composables/useReadoutStatus'

function setup(initial: ReadoutStatus = 'running') {
  vi.useFakeTimers()
  const input = ref(initial)
  const scope = effectScope()
  const shown = scope.run(() => useReadoutStatus(input, { appearAfter: 220, holdFor: 450 }))!
  return { input, shown, stop: () => scope.stop() }
}

afterEach(() => vi.useRealTimers())

describe('readout captions', () => {
  it('suppresses transient quiet and lock captions without delaying the input', () => {
    const s = setup()
    for (const next of ['tooQuiet', 'locked', 'recorded'] as const) {
      s.input.value = next
      expect(s.input.value).toBe(next)
      vi.advanceTimersByTime(150)
      s.input.value = 'running'
      vi.advanceTimersByTime(500)
      expect(s.shown.value).toBe('running')
    }
    s.stop()
  })

  it('shows a persistent advisory and holds it through a short threshold crossing', () => {
    const s = setup()
    s.input.value = 'tooQuiet'
    vi.advanceTimersByTime(220)
    expect(s.shown.value).toBe('tooQuiet')
    s.input.value = 'running'
    vi.advanceTimersByTime(449)
    expect(s.shown.value).toBe('tooQuiet')
    vi.advanceTimersByTime(1)
    expect(s.shown.value).toBe('running')
    s.stop()
  })

  it('never keeps a lock or warm-up instruction after its source state ends', () => {
    for (const state of ['initializing', 'locked', 'recorded'] as const) {
      const s = setup(state)
      vi.advanceTimersByTime(220)
      expect(s.shown.value).toBe(state)
      s.input.value = 'running'
      expect(s.shown.value).toBe('running')
      s.stop()
    }
  })

  it('shows stop, device loss and explicit freeze immediately and cancels pending captions', () => {
    const s = setup()
    for (const state of ['stopped', 'deviceLost', 'frozen'] as const) {
      s.input.value = 'locked'
      vi.advanceTimersByTime(100)
      s.input.value = state
      expect(s.shown.value).toBe(state)
      vi.advanceTimersByTime(1000)
      expect(s.shown.value).toBe(state)
    }
    s.stop()
  })

  it('cancels on disposal and requires a continuous stable candidate', () => {
    const s = setup()
    s.input.value = 'locked'
    vi.advanceTimersByTime(200)
    s.input.value = 'running'
    s.input.value = 'locked'
    vi.advanceTimersByTime(200)
    expect(s.shown.value).toBe('running')
    s.stop()
    vi.advanceTimersByTime(1000)
    expect(s.shown.value).toBe('running')
    expect(vi.getTimerCount()).toBe(0)
  })
})
