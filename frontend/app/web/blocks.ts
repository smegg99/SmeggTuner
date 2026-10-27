// app/web/blocks.ts
import type { PerfMeter } from './perf'

// A block older than this when the worker reaches it is dropped: a late reading is worse than a missing one.
const STALE_MS = 250

interface Clock {
  wall: () => number
  mono: () => number
}

// createBlockSink feeds captured blocks to the engine; at is Date.now() when the worklet captured the block.
export function createBlockSink(
  push: (samples: Float32Array, lost: number) => void,
  perf: PerfMeter,
  rate: number,
  clock: Clock = { wall: () => Date.now(), mono: () => performance.now() },
) {
  let lost = 0
  return (samples: Float32Array, at: number) => {
    if (clock.wall() - at > STALE_MS) {
      lost += samples.length
      perf.drop()
      return
    }
    const t0 = clock.mono()
    push(samples, lost)
    lost = 0
    const t1 = clock.mono()
    perf.observe(t1 - t0, samples.length / rate * 1000, t1)
  }
}
