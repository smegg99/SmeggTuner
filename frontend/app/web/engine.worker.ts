// app/web/engine.worker.ts
// The Go engine in a dedicated worker: answers binding calls, takes microphone blocks, reports its own load.
import { createBlockSink } from './blocks'
import { PerfMeter } from './perf'
import { PERF_EVENT, type ToWorker } from './protocol'

interface GoApi {
  call(methodID: number, args: string): string
  push(samples: Float32Array, lost: number): void
}
interface GoRuntime {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

const scope = globalThis as unknown as {
  postMessage(msg: unknown): void
  onmessage: ((e: MessageEvent) => void) | null
  Go: new () => GoRuntime
  smeggtuner?: GoApi
  smeggtunerInit?: { config: string, rate: number }
  smeggtunerPost?: (json: string) => void
  smeggtunerReady?: () => void
}

const perf = new PerfMeter()
let sink: ((samples: Float32Array, at: number) => void) | null = null

async function fetchWithProgress(url: string): Promise<ArrayBuffer> {
  const res = await fetch(url)
  if (!res.ok || !res.body) throw new Error(`${url}: HTTP ${res.status}`)
  const total = Number(res.headers.get('content-length')) || 0
  const reader = res.body.getReader()
  const parts: Uint8Array[] = []
  let loaded = 0
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    parts.push(value)
    loaded += value.length
    scope.postMessage({ type: 'progress', loaded, total })
  }
  const out = new Uint8Array(loaded)
  let at = 0
  for (const p of parts) {
    out.set(p, at)
    at += p.length
  }
  return out.buffer
}

async function boot(m: Extract<ToWorker, { type: 'init' }>) {
  try {
    await import(/* @vite-ignore */ new URL('wasm_exec.js', new URL(m.base, location.origin)).href)
    scope.smeggtunerInit = { config: m.config, rate: m.rate }
    scope.smeggtunerPost = json => scope.postMessage(JSON.parse(json))
    const ready = new Promise<void>(resolve => (scope.smeggtunerReady = resolve))
    const go = new scope.Go()
    const bytes = await fetchWithProgress(new URL('smeggtuner.wasm', new URL(m.base, location.origin)).href)
    const { instance } = await WebAssembly.instantiate(bytes, go.importObject)
    void go.run(instance)
    await ready
    const api = scope.smeggtuner!
    sink = createBlockSink((s, lost) => api.push(s, lost), perf, m.rate)
    setInterval(() => scope.postMessage({ type: 'event', name: PERF_EVENT, data: perf.snapshot(performance.now()) }), 1000)
    scope.postMessage({ type: 'ready' })
  }
  catch (e) {
    scope.postMessage({ type: 'failed', message: String(e) })
  }
}

scope.onmessage = (e: MessageEvent) => {
  const m = e.data as ToWorker | { type: 'audio', port: MessagePort }
  switch (m.type) {
    case 'init':
      void boot(m)
      break
    case 'call':
      scope.postMessage({ type: 'result', id: m.id, reply: scope.smeggtuner?.call(m.methodID, m.args)
        ?? '{"ok":false,"kind":"RuntimeError","message":"engine not started"}' })
      break
    case 'audio':
      m.port.onmessage = (b: MessageEvent<{ samples: Float32Array, at: number }>) => sink?.(b.data.samples, b.data.at)
      break
  }
}
