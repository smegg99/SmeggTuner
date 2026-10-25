import { alpha, cssColor } from '~/composables/canvasColor'
import { BAND_SPECTRUM_COLUMNS } from '~/composables/tuner/tunerProtocol'
import type { Geometry } from './spectrumGeometry'

// One small spectrum per chord tone or register band, side by side, each across +-50 cents.

export interface Pane {
  left: number
  width: number
}

/** paneLayout splits [left, left+width] into count equal panes with gap between them. */
export function paneLayout(left: number, width: number, count: number, gap: number): Pane[] {
  if (count <= 0) return []
  const w = Math.max(0, (width - gap * (count - 1)) / count)
  return Array.from({ length: count }, (_, i) => ({ left: left + i * (w + gap), width: w }))
}

export function drawPanes(
  ctx: CanvasRenderingContext2D,
  g: Geometry,
  font: string,
  colors: Record<string, unknown>,
  pictures: readonly Float32Array[],
  labels: readonly string[],
) {
  const accent = cssColor(colors.accent)
  const height = g.base - g.top
  const panes = paneLayout(g.padL, g.inner, pictures.length, g.fs * 1.2)

  ctx.font = `${g.fs}px ${font}`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'top'

  panes.forEach((pane, k) => {
    const xOf = (i: number) => pane.left + (i / (BAND_SPECTRUM_COLUMNS - 1)) * pane.width
    const yOf = (v: number) => g.base - Math.max(0, Math.min(1, v)) * height
    const picture = pictures[k]!
    const mid = Math.round(pane.left + pane.width / 2) + 0.5

    ctx.strokeStyle = cssColor(colors.ink3)
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(mid, g.top)
    ctx.lineTo(mid, g.base)
    ctx.stroke()

    const path = new Path2D()
    path.moveTo(pane.left, g.base)
    for (let i = 0; i < BAND_SPECTRUM_COLUMNS; i++) path.lineTo(xOf(i), yOf(picture[i] ?? 0))
    path.lineTo(pane.left + pane.width, g.base)
    path.closePath()
    const fill = ctx.createLinearGradient(0, g.top, 0, g.base)
    fill.addColorStop(0, alpha(accent, 0.34))
    fill.addColorStop(1, alpha(accent, 0.02))
    ctx.fillStyle = fill
    ctx.fill(path)

    ctx.strokeStyle = accent
    ctx.lineWidth = 1.6
    ctx.beginPath()
    for (let i = 0; i < BAND_SPECTRUM_COLUMNS; i++) {
      if (i === 0) ctx.moveTo(xOf(i), yOf(picture[i] ?? 0))
      else ctx.lineTo(xOf(i), yOf(picture[i] ?? 0))
    }
    ctx.stroke()

    ctx.strokeStyle = cssColor(colors.neutral)
    ctx.globalAlpha = 0.35
    ctx.beginPath()
    ctx.moveTo(pane.left, Math.round(g.base) + 0.5)
    ctx.lineTo(pane.left + pane.width, Math.round(g.base) + 0.5)
    ctx.stroke()
    ctx.globalAlpha = 1

    ctx.fillStyle = cssColor(colors.ink3)
    ctx.fillText(labels[k] || '0', mid, g.base + g.lane * 0.35)
  })
}
