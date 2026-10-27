// tests/web/perf.spec.ts
import { describe, expect, it } from 'vitest'
import { PerfMeter } from '~/web/perf'

// Feeds blocks of 20 ms audio at real time, each costing work ms, for seconds of wall time.
function feed(p: PerfMeter, from: number, seconds: number, work: number) {
  let now = from
  for (; now < from + seconds * 1000; now += 20) p.observe(work, 20, now)
  return now
}

describe('PerfMeter', () => {
  it('reports load as work over audio', () => {
    const p = new PerfMeter()
    const now = feed(p, 0, 2, 5)
    expect(p.snapshot(now).load).toBeCloseTo(0.25, 2)
    expect(p.snapshot(now).behind).toBe(false)
  })

  it('is behind after sustained load', () => {
    const p = new PerfMeter()
    let now = feed(p, 0, 2, 19)
    expect(p.snapshot(now).behind).toBe(false)
    now = feed(p, now, 2, 19)
    expect(p.snapshot(now).behind).toBe(true)
  })

  it('recovers when the load drops', () => {
    const p = new PerfMeter()
    let now = feed(p, 0, 4, 19)
    now = feed(p, now, 6, 2)
    expect(p.snapshot(now).behind).toBe(false)
  })

  it('counts dropped blocks', () => {
    const p = new PerfMeter()
    p.drop()
    p.drop()
    expect(p.snapshot(0).dropped).toBe(2)
  })
})
