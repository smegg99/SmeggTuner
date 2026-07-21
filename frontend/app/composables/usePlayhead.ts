import { computed, ref } from 'vue'

// The needle coasts on the wall clock between Go's ~30/s reports; drift is corrected by nudging the rate, never the position (which would jitter backwards). Past SNAP it is a seek or loop, so jump there.

const base = ref(0)
let baseAt = 0
let rate = 1

/* Explicit, not `baseAt === 0`: performance.now() really is 0 on page load. */
let seeded = false

const reported = ref(0)

/** Past this much, the needle was moved (seek, loop, stop) rather than drifted. */
const SNAP = 0.1

/** How hard the rate leans on a disagreement. 30 ms out puts the clock ~6% off nominal. */
const GAIN = 2

/** The clock may run this far from real time, and no further. Well below what an eye sees. */
const RATE_MIN = 0.85
const RATE_MAX = 1.15

/** Max coast before the needle freezes: if events stop while the engine claims to run, it must not run off on a guess. */
const MAX_COAST = 0.25

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

export function usePlayhead() {
  const at = (advancing: boolean, limit: number): number => {
    if (!advancing || !seeded) return base.value

    const elapsed = Math.min((performance.now() - baseAt) / 1000, MAX_COAST)
    return Math.min(base.value + elapsed * rate, limit)
  }

  /** Report a position from Go; `authoritative` marks a command reply (seek/pause/selection) to jump to rather than ease toward. */
  const land = (seconds: number, advancing: boolean, limit: number, authoritative = false) => {
    reported.value = seconds

    const now = performance.now()
    const drawn = at(advancing, limit)

    if (!advancing || authoritative || !seeded) {
      base.value = seconds
      baseAt = now
      rate = 1
      seeded = true
      return
    }

    const err = seconds - drawn

    // Moved, not drifted.
    if (Math.abs(err) > SNAP) {
      base.value = seconds
      baseAt = now
      rate = 1
      return
    }

    base.value = drawn
    baseAt = now
    rate = clamp(1 + err * GAIN, RATE_MIN, RATE_MAX)
  }

  const heard = computed(() => reported.value)

  return { land, at, heard }
}
