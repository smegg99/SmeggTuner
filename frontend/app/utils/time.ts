// Timeline ticks land on round numbers a person would say, never the window divided by ten (0.37s, 0.74s...) — hence a table, not a division.
export interface TimeTick {
  at: number
  major: boolean
  label: string
}

/** The steps a clock is allowed to count in, from a millisecond up to a minute. */
const STEPS = [
  0.001, 0.002, 0.005,
  0.01, 0.02, 0.05,
  0.1, 0.2, 0.5,
  1, 2, 5,
  10, 15, 30,
  60, 120, 300,
]

/** Below this many pixels apart, labels collide and the ruler is worse than none. */
const MIN_LABEL_PX = 64

/** formatTime is m:ss for a real duration, and seconds with decimals once zoomed in. */
export function formatTime(at: number, step: number): string {
  if (step >= 1) {
    const m = Math.floor(at / 60)
    const s = Math.floor(at % 60)
    return m > 0 ? `${m}:${String(s).padStart(2, '0')}` : `${s}s`
  }

  // enough decimals to tell one tick from the next, and not one more
  const places = step >= 0.1 ? 1 : step >= 0.01 ? 2 : 3
  return `${at.toFixed(places)}s`
}

// timeTicks lays out the ruler between two times across a width; majors carry a label and gridline, minors are the five unlabelled marks between.
export function timeTicks(from: number, to: number, width: number): TimeTick[] {
  const span = to - from
  if (!(span > 0) || width <= 0) return []

  // the first step whose labels are far enough apart to be read
  const step = STEPS.find(s => (s / span) * width >= MIN_LABEL_PX) ?? STEPS[STEPS.length - 1]!
  const minor = step / 5

  const out: TimeTick[] = []
  const first = Math.ceil(from / minor) * minor

  for (let at = first; at <= to; at += minor) {
    // Floating point: 0.1*3 is 0.30000000000000004; snap before testing against a step multiple, which decides the label.
    const t = Math.round(at / minor) * minor
    const major = Math.abs(Math.round(t / step) * step - t) < minor / 2

    out.push({ at: t, major, label: major ? formatTime(t, step) : '' })
  }
  return out
}
