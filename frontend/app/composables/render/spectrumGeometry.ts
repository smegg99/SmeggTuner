import type { CanvasView } from '~/composables/useCanvas'
import { centToFrac } from '~/utils/scale'

export interface Geometry {
  padX: number
  /** left edge of the plot */
  padL: number
  top: number
  base: number
  lane: number
  inner: number
  fs: number
  xOf: (cent: number) => number
}

// The plot runs edge to edge; the dB numbers sit inside it at the left, on their gridlines.
export function geometry(view: CanvasView): Geometry {
  const padX = 0
  const fs = Math.max(9, Math.min(13, view.height * 0.075))
  const padL = 0
  const lane = Math.max(12, Math.min(26, view.height * 0.15))
  const inner = view.width - padL - padX
  const top = 2 // the 0 dB label hangs under the top line

  return {
    padX,
    padL,
    lane,
    inner,
    fs,
    top,
    base: view.height - lane,
    xOf: (cent: number) => padL + centToFrac(cent) * inner,
  }
}
