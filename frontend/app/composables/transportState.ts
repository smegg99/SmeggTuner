// Leaf module: never imports the transport composable, so the mirror side can depend on it without a cycle.
import { reactive } from 'vue'
import type { TransportDTO } from '~~bindings/smegg.me/smeggtuner/services/audio/models.js'

/** The transport as Go last reported it. Seconds throughout. */
export const state = reactive<TransportDTO>({
  available: false,
  name: '',
  duration: 0,
  position: 0,
  from: 0,
  to: 0,
  paused: false,
  moving: false,
  loop: false,
  sampleRate: 0,
})

/** The zoom window: the stretch of file the canvas shows. The view's own state; Go
 *  is never told about it. */
export const view = reactive({ from: 0, to: 0 })

/** Narrowest window we zoom to: below a millisecond the peaks go flat. */
const MIN_SPAN = 0.001

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

/** Show the whole file. */
export function fit() {
  view.from = 0
  view.to = state.duration
}

/** Show the selection. */
export function fitSelection() {
  if (state.to <= state.from) return
  view.from = state.from
  view.to = state.to
}

// Keep the anchor (where the pointer is) fixed and scale the window around it, then
// push it back inside the file so no zoom shows empty space past the ends.
export function zoom(factor: number, anchor: number) {
  if (!state.available) return

  const span = clamp((view.to - view.from) * factor, MIN_SPAN, state.duration)
  const at = clamp(anchor, view.from, view.to)
  const share = (at - view.from) / Math.max(view.to - view.from, 1e-9)

  let from = at - span * share
  let to = from + span

  if (from < 0) [from, to] = [0, span]
  if (to > state.duration) [from, to] = [state.duration - span, state.duration]

  view.from = Math.max(0, from)
  view.to = Math.min(state.duration, to)
}

/** Slide the window without changing its width. */
export function pan(seconds: number) {
  const span = view.to - view.from
  const from = clamp(view.from + seconds, 0, Math.max(0, state.duration - span))

  view.from = from
  view.to = from + span
}

/** Put the window's left edge here: what dragging the overview's box does. */
export function scrollTo(from: number) {
  pan(from - view.from)
}
