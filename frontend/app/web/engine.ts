// app/web/engine.ts
// Boots the engine worker and wires it to the generated bindings, the Wails event bus and the microphone.
import '@wailsio/runtime'
import workletUrl from './capture.worklet.ts?worker&url'
import { loadConfig, saveConfig } from './storage'
import { installWorkerTransport } from './transport'
import type { FromWorker } from './protocol'

export const ENGINE_BASE = '/engine/'

export interface WebEngine {
  context: AudioContext
  rate: number
  attach(stream: MediaStream): void
  detach(): void
}

type Dispatch = (e: { name: string, data: unknown }) => void

export async function bootEngine(
  onProgress: (loaded: number, total: number) => void,
  local: Map<number, (args: unknown[]) => unknown>,
): Promise<WebEngine> {
  // Created before any gesture so its rate is known; it stays suspended until the Start tap resumes it.
  const context = new AudioContext({ latencyHint: 'interactive' })
  const worker = new Worker(new URL('./engine.worker.ts', import.meta.url), { type: 'module' })
  installWorkerTransport(worker, local)
  const dispatch = (window as unknown as { _wails: { dispatchWailsEvent: Dispatch } })._wails.dispatchWailsEvent

  const ready = new Promise<void>((resolve, reject) => {
    worker.addEventListener('message', (e: MessageEvent<FromWorker>) => {
      const m = e.data
      if (m.type === 'event') dispatch({ name: m.name, data: m.data })
      else if (m.type === 'config') saveConfig(m.json)
      else if (m.type === 'progress') onProgress(m.loaded, m.total)
      else if (m.type === 'ready') resolve()
      else if (m.type === 'failed') reject(new Error(m.message))
    })
  })
  worker.postMessage({ type: 'init', config: loadConfig(), rate: context.sampleRate, base: ENGINE_BASE })
  await ready

  await context.audioWorklet.addModule(workletUrl)
  const node = new AudioWorkletNode(context, 'smeggtuner-capture', { numberOfInputs: 1, numberOfOutputs: 0 })
  const channel = new MessageChannel()
  node.port.postMessage({ type: 'port', port: channel.port1 }, [channel.port1])
  worker.postMessage({ type: 'audio', port: channel.port2 }, [channel.port2])

  let source: MediaStreamAudioSourceNode | null = null
  return {
    context,
    rate: context.sampleRate,
    attach(stream) {
      source?.disconnect()
      source = context.createMediaStreamSource(stream)
      source.connect(node)
    },
    detach() {
      source?.disconnect()
      source = null
    },
  }
}
