import type { CanvasView } from '~/composables/useCanvas'
import { centToFrac } from '~/utils/scale'

export interface Geometry {
  padX: number
  /** left edge of the plot, past the decibel gutter */
  padL: number
  top: number
  base: number
  lane: number
  inner: number
  fs: number
  xOf: (cent: number) => number
}

// The dB numbers get their own gutter; the plot begins to its right so no number is clipped.
export function geometry(view: CanvasView): Geometry {
  const padX = view.width * 0.018
  const fs = Math.max(9, Math.min(13, view.height * 0.075))
  const gutter = fs * 2.3 // wide enough for "-60"
  const padL = padX + gutter
  const lane = Math.max(12, Math.min(26, view.height * 0.15))
  const inner = view.width - padL - padX
  const top = Math.max(view.height * 0.07, fs * 0.8) // room for the 0 dB label to sit on its line

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
