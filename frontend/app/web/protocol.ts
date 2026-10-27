// app/web/protocol.ts
// Messages between the page and the engine worker; the Go side is backend/web.

export interface EnginePort {
  postMessage(msg: unknown, transfer?: Transferable[]): void
  addEventListener(type: 'message', fn: (e: MessageEvent) => void): void
}

export type ToWorker
  = | { type: 'init', config: string, rate: number, base: string }
    | { type: 'call', id: number, methodID: number, args: string }

export type FromWorker
  = | { type: 'progress', loaded: number, total: number }
    | { type: 'ready' }
    | { type: 'failed', message: string }
    | { type: 'result', id: number, reply: string }
    | { type: 'event', name: string, data: unknown }
    | { type: 'config', json: string }

// Reply is the JSON backend/web.callJSON answers one binding call with.
export interface Reply {
  ok: boolean
  data?: unknown
  kind?: string
  message?: string
  cause?: unknown
}

// PerfEvent is the worker's own event: engine load over the last seconds, blocks dropped as stale, and whether it cannot keep up.
export const PERF_EVENT = 'web:perf'
export interface PerfDTO {
  load: number
  dropped: number
  behind: boolean
}
