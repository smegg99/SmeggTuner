// One shared requestAnimationFrame for the whole app. It stops when every painter reports nothing moving,
// and is woken again by anything that dirties a picture (a measurement, a theme change, a resize).

export interface Painter {
  /** dt is seconds since the last tick. Returns true while there is more to do. */
  tick: (dt: number) => boolean
}

// Cap the first dt after a backgrounded window so interpolators ease in instead of teleporting.
const MAX_DT = 0.1

// Cap at 60fps: on WebKitGTK a higher rate redraws everything for a picture the eye can't tell apart and drops frames.
// The minus one is slack, since rAF does not fire on exact multiples.
const MIN_FRAME_MS = 1000 / 60 - 1

const painters = new Set<Painter>()

let raf = 0
let last = 0

function frame(now: number) {
  if (last !== 0 && now - last < MIN_FRAME_MS) {
    raf = requestAnimationFrame(frame)
    return
  }

  const dt = last === 0 ? 0 : Math.min(MAX_DT, (now - last) / 1000)
  last = now

  let busy = false
  for (const painter of painters) {
    if (painter.tick(dt)) busy = true
  }

  if (busy) {
    raf = requestAnimationFrame(frame)
    return
  }

  // Settled. The next dt must not be measured from the frame the loop stopped
  // on, which may be minutes ago.
  raf = 0
  last = 0
}

/** wake runs the loop again if it has settled. Cheap enough to call per event. */
export function wakeFrameLoop() {
  if (raf === 0) raf = requestAnimationFrame(frame)
}

export function addPainter(painter: Painter) {
  painters.add(painter)
  wakeFrameLoop()
}

export function removePainter(painter: Painter) {
  painters.delete(painter)

  if (painters.size === 0 && raf !== 0) {
    cancelAnimationFrame(raf)
    raf = 0
    last = 0
  }
}
