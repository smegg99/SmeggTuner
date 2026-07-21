import type { CanvasView } from '~/composables/useCanvas'
import { centToFrac } from '~/utils/scale'

export interface Geometry {
  padX: number
  xOf: (cent: number) => number
  labelY: number
  scaleTop: number
  scaleBot: number
  hzY: number
  fs: number
  hzFs: number
  needle: number
}

// Every size is a fraction of the panel so proportions hold at any size.
export function geometry(view: CanvasView): Geometry {
  const padX = view.width * 0.024
  const inner = view.width - padX * 2

  const fs = Math.max(9, Math.min(26, view.height * 0.062))
  const hzFs = Math.max(8, Math.min(20, view.height * 0.055))

  // Floor the label lanes at half a glyph so short panels don't clip the top row / reading box.
  const labelY = Math.max(view.height * 0.07, fs * 0.5 + 2)
  const hzY = view.height - Math.max(view.height * 0.08, hzFs * 0.8 + 3)

  return {
    padX,
    xOf: (cent: number) => padX + centToFrac(cent) * inner,
    labelY,
    scaleTop: view.height * 0.145,
    scaleBot: hzY - view.height * 0.085,
    hzY,
    fs,
    hzFs,
    needle: Math.max(1.5, Math.min(5, view.height * 0.011)),
  }
}
