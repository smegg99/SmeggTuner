// tests/setup.ts
import { afterAll, afterEach, expect, vi } from 'vitest'
import { ref } from 'vue'

const unexpected: string[] = []

// Keep the real runtime; only its optional bootstrap script is absent from this host.
const isolatedFetch = vi.fn<typeof fetch>(async (input, init) => {
  const request = input instanceof Request ? input : undefined
  const page = typeof window === 'undefined' ? undefined : window.location.href
  const url = new URL(request?.url ?? String(input), page ?? 'http://unit.test/')
  const method = (init?.method ?? request?.method ?? 'GET').toUpperCase()
  if (page && method === 'HEAD' && url.href === new URL('/wails/custom.js', page).href) {
    return new Response(null, { status: 404 })
  }

  const message = `Unexpected network request in unit test: ${method} ${url.href}`
  unexpected.push(message)
  throw new Error(message)
})

// This is the host baseline, so per-test unstubAllGlobals restores it instead of real fetch.
Object.defineProperty(globalThis, 'fetch', { configurable: true, writable: true, value: isolatedFetch })

// A caller catching the rejection must not turn an unexpected request into a passing test.
const checkNetwork = () => expect(unexpected.splice(0), 'Unexpected unit-test network access').toEqual([])
afterEach(checkNetwork)
afterAll(checkNetwork)

// Component tests use the desktop layout unless a case supplies a compact viewport.
Object.defineProperty(globalThis, 'useCompactLayout', { configurable: true, writable: true, value: () => ref(false) })
Object.defineProperty(globalThis, 'usePagedLayout', { configurable: true, writable: true, value: () => ref(false) })
