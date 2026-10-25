import type { CanvasView } from '~/composables/useCanvas'
import { NOTE_BANDS } from '~/composables/useTuner'

const SHARPS = new Set([1, 3, 6, 8, 10])
export const isSharp = (midi: number) => SHARPS.has(midi % 12)

export interface Geometry {
  padX: number
  lane: number
  base: number
  top: number
  bw: number
  xOf: (band: number) => number
  fs: number
}

// The bands run edge to edge; the names are clamped inside the panel where they are drawn.
export function geometry(view: CanvasView): Geometry {
  const padX = 0
  const lane = Math.max(16, Math.min(42, view.height * 0.38))
  const bw = (view.width - padX * 2) / NOTE_BANDS

  return {
    padX,
    lane,
    base: view.height - lane,
    top: 2,
    bw,
    xOf: (band: number) => padX + band * bw,
    fs: Math.max(8, Math.min(17, view.height * 0.155)),
  }
}

// Edge to edge like geometry(), or edge clicks land on the wrong band.
export function bandAt(clientX: number, rect: DOMRect): number {
  const bw = rect.width / NOTE_BANDS
  const band = Math.floor((clientX - rect.left) / bw)

  return band >= 0 && band < NOTE_BANDS ? band : -1
}
