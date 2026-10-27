// tests/web/blocks.spec.ts
import { describe, expect, it } from 'vitest'
import { createBlockSink } from '~/web/blocks'
import { PerfMeter } from '~/web/perf'

function rig() {
  const pushed: { n: number, lost: number }[] = []
  const wall = 1000
  let mono = 0
  const perf = new PerfMeter()
  const push = (s: Float32Array, lost: number) => {
    pushed.push({ n: s.length, lost })
    mono += 3
  }
  const sink = createBlockSink(push, perf, 48000, {
    wall: () => wall,
    mono: () => mono,
  })
  return { sink, pushed, perf }
}

describe('worker block sink', () => {
  it('pushes a fresh block and measures its cost', () => {
    const { sink, pushed, perf } = rig()
    sink(new Float32Array(960), 990)
    expect(pushed).toEqual([{ n: 960, lost: 0 }])
    expect(perf.snapshot(0).load).toBeCloseTo(3 / 20, 3)
  })

  it('drops a stale block and charges its frames to the next push', () => {
    const { sink, pushed, perf } = rig()
    sink(new Float32Array(960), 500)
    expect(pushed).toEqual([])
    sink(new Float32Array(960), 1000)
    expect(pushed).toEqual([{ n: 960, lost: 960 }])
    expect(perf.snapshot(0).dropped).toBe(1)
  })
})
