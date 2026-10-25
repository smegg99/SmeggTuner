import type { CanvasView } from '~/composables/useCanvas'
import { centToFrac } from '~/utils/scale'

export interface Geometry {
  padX: number
  xOf: (cent: number) => number
  labelY: number
  scaleTop: number
  scaleBot: number
  fs: number
}

// A thin strip: the numbers along the top, the scale and needles filling the rest. The app's 4px gap
// above the numbers and below the scale, and room at the ends for the 50s, so nothing touches the frame.
const GAP = 4

export function geometry(view: CanvasView): Geometry {
  const fs = Math.max(9, Math.min(13, view.height * 0.24))
  const padX = Math.round(fs * 1.5)
  const inner = view.width - padX * 2

  return {
    padX,
    xOf: (cent: number) => padX + centToFrac(cent) * inner,
    labelY: GAP + fs / 2,
    scaleTop: GAP + fs + GAP,
    scaleBot: view.height - GAP,
    fs,
  }
}
