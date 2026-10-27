// tests/web/sw.spec.ts
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it, vi } from 'vitest'

const source = readFileSync(path.resolve(__dirname, '../../web-public/sw.js'), 'utf8')

// Runs sw.js against a fake worker scope and cache store; build is what scripts/web.sh stamps in.
function load(build: string, existing: string[]) {
  const handlers: Record<string, (e: unknown) => void> = {}
  const names = new Set(existing)
  const caches = {
    keys: async () => [...names],
    delete: vi.fn(async (n: string) => names.delete(n)),
    open: async (n: string) => {
      names.add(n)
      return { addAll: async () => {}, put: async () => {}, match: async () => undefined }
    },
    match: async () => undefined,
  }
  const self = {
    addEventListener: (t: string, fn: (e: unknown) => void) => {
      handlers[t] = fn
    },
    skipWaiting: () => {},
    clients: { claim: async () => {} },
    location: { origin: 'https://app.smeggtuner.com' },
  }
  new Function('self', 'caches', 'fetch', source.replaceAll('__BUILD__', build))(self, caches, vi.fn())
  return { handlers, names, caches }
}

async function run(handler: (e: unknown) => void) {
  let done: Promise<unknown> = Promise.resolve()
  handler({
    waitUntil: (p: Promise<unknown>) => {
      done = p
    },
  })
  await done
}

describe('service worker', () => {
  it('cache name carries version and hash', async () => {
    const { handlers, names } = load('1.4.0-abc123', [])
    await run(handlers.install!)
    expect([...names]).toEqual(['smeggtuner-web-1.4.0-abc123'])
  })

  it('activate deletes older caches and keeps other origins\' data', async () => {
    const { handlers, names } = load('1.4.0-new', ['smeggtuner-web-1.3.0-old', 'smeggtuner-web-1.4.0-new', 'unrelated'])
    await run(handlers.activate!)
    expect([...names].sort()).toEqual(['smeggtuner-web-1.4.0-new', 'unrelated'])
  })
})
