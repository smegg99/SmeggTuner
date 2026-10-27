// app/web/transport.ts
import { setTransport } from '@wailsio/runtime'
import type { EnginePort, FromWorker, Reply } from './protocol'

// The runtime's object and method numbers for a bound-method call (runtime.js objectNames.Call, calls.js CallBinding).
const CALL_OBJECT = 0
const CALL_BINDING = 0

class RuntimeError extends Error {
  override name = 'RuntimeError'
}

function errorOf(r: Reply): Error {
  const message = r.message ?? 'call failed'
  const err = r.kind === 'TypeError'
    ? new TypeError(message)
    : r.kind === 'ReferenceError' ? new ReferenceError(message) : new RuntimeError(message)
  ;(err as Error & { cause?: unknown }).cause = r.cause
  return err
}

// createWorkerTransport answers the generated bindings from the engine worker instead of the desktop app's HTTP endpoint;
// methods in local are answered on the page.
export function createWorkerTransport(port: EnginePort, local: Map<number, (args: unknown[]) => unknown> = new Map()) {
  let next = 0
  const pending = new Map<number, { resolve: (v: unknown) => void, reject: (e: Error) => void }>()
  port.addEventListener('message', (e: MessageEvent<FromWorker>) => {
    if (e.data?.type !== 'result') return
    const p = pending.get(e.data.id)
    if (!p) return
    pending.delete(e.data.id)
    const r = JSON.parse(e.data.reply) as Reply
    if (r.ok) p.resolve(r.data ?? null)
    else p.reject(errorOf(r))
  })
  return {
    // Window, Browser, System and the other runtime objects have no worker side; the web shell does not depend on them.
    call(objectID: number, method: number, _window: string, args: { methodID?: number, args?: unknown[] }): Promise<unknown> {
      if (objectID !== CALL_OBJECT || method !== CALL_BINDING || args?.methodID === undefined) return Promise.resolve(undefined)
      const handler = local.get(args.methodID)
      if (handler) return Promise.resolve().then(() => handler(args.args ?? []))
      const id = ++next
      return new Promise((resolve, reject) => {
        pending.set(id, { resolve, reject })
        port.postMessage({ type: 'call', id, methodID: args.methodID, args: JSON.stringify(args.args ?? []) })
      })
    },
  }
}

export function installWorkerTransport(port: EnginePort, local: Map<number, (args: unknown[]) => unknown>) {
  setTransport(createWorkerTransport(port, local))
}
