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

// gutter is a name's width so the first and last C names stay inside the panel.
export function geometry(view: CanvasView, gutter: number): Geometry {
  const padX = gutter
  const lane = Math.max(16, Math.min(42, view.height * 0.38))
  const bw = (view.width - padX * 2) / NOTE_BANDS

  return {
    padX,
    lane,
    base: view.height - lane,
    top: view.height * 0.08,
    bw,
    xOf: (band: number) => padX + band * bw,
    fs: Math.max(8, Math.min(17, view.height * 0.155)),
  }
}

// GUTTER and bandAt must use the same padding as geometry(), or edge clicks land on the wrong band.
export const GUTTER = 0.022

export function bandAt(clientX: number, rect: DOMRect): number {
  const padX = rect.width * GUTTER
  const bw = (rect.width - padX * 2) / NOTE_BANDS
  const band = Math.floor((clientX - rect.left - padX) / bw)

  return band >= 0 && band < NOTE_BANDS ? band : -1
}
