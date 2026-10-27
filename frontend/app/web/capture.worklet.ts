// app/web/capture.worklet.ts
// Runs on the audio thread: gathers the 128-frame render quanta of channel 0 into engine-sized blocks.
declare class AudioWorkletProcessor {
  readonly port: MessagePort
}
declare function registerProcessor(name: string, ctor: new () => AudioWorkletProcessor): void

const BLOCK = 1024

class Capture extends AudioWorkletProcessor {
  private out: MessagePort | null = null
  private buf = new Float32Array(BLOCK)
  private n = 0

  constructor() {
    super()
    this.port.onmessage = (e: MessageEvent<{ type: string, port: MessagePort }>) => {
      if (e.data?.type === 'port') this.out = e.data.port
    }
  }

  process(inputs: Float32Array[][]): boolean {
    const ch = inputs[0]?.[0]
    if (!ch || !this.out) return true
    for (let i = 0; i < ch.length; i++) {
      this.buf[this.n++] = ch[i]!
      if (this.n === BLOCK) {
        this.out.postMessage({ samples: this.buf, at: Date.now() }, [this.buf.buffer])
        this.buf = new Float32Array(BLOCK)
        this.n = 0
      }
    }
    return true
  }
}

registerProcessor('smeggtuner-capture', Capture)
