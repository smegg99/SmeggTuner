// tests/web/fakePort.ts
import type { EnginePort } from '~/web/protocol'

// A worker stand-in: records what the page posts and lets a test answer as the worker would.
export function fakePort() {
  const posted: unknown[] = []
  const listeners: ((e: MessageEvent) => void)[] = []
  const port: EnginePort = {
    postMessage: (msg: unknown) => { posted.push(msg) },
    addEventListener: (_type: 'message', fn: (e: MessageEvent) => void) => { listeners.push(fn) },
  }
  const reply = (data: unknown) => listeners.forEach(fn => fn({ data } as MessageEvent))
  return { port, posted, reply }
}
