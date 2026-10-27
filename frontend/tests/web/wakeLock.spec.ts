// tests/web/wakeLock.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { useWakeLock } from '~/web/wakeLock'

function stubLock() {
  const sentinel = {
    released: false,
    release: vi.fn(async () => {
      sentinel.released = true
    }),
  }
  const request = vi.fn(async () => {
    sentinel.released = false
    return sentinel
  })
  vi.stubGlobal('navigator', { ...navigator, wakeLock: { request } })
  return { request, sentinel }
}

afterEach(() => vi.unstubAllGlobals())

async function flush() {
  await nextTick()
  await Promise.resolve()
  await Promise.resolve()
}

describe('useWakeLock', () => {
  it('requests on start and releases on stop', async () => {
    const { request, sentinel } = stubLock()
    const running = ref(false)
    useWakeLock(running)
    running.value = true
    await flush()
    expect(request).toHaveBeenCalledWith('screen')
    running.value = false
    await flush()
    expect(sentinel.release).toHaveBeenCalled()
  })

  it('re-requests when the tab becomes visible', async () => {
    const { request, sentinel } = stubLock()
    const running = ref(true)
    useWakeLock(running)
    await flush()
    sentinel.released = true
    document.dispatchEvent(new Event('visibilitychange'))
    await flush()
    expect(request).toHaveBeenCalledTimes(2)
  })

  it('is a no-op without the API', async () => {
    vi.stubGlobal('navigator', { ...navigator, wakeLock: undefined })
    const running = ref(false)
    useWakeLock(running)
    running.value = true
    await flush()
  })
})
