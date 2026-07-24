import { useTheme } from 'vuetify'
import { alpha, cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { SPECTRUM_CENTS, SPECTRUM_COLUMNS, SPECTRUM_FLOOR_DB, live, useTuner } from '~/composables/useTuner'
import { geometry, type Geometry } from './spectrumGeometry'

// The backend already bins the spectrum into these columns, so there is no binning here.

// Column height is 1 + dB/SPECTRUM_FLOOR_DB (0 dB top, -60 floor). Must match core/dsp.spectrumHeight.
function drawDecibels(ctx: CanvasRenderingContext2D, view: CanvasView, g: Geometry, c: Record<string, unknown>) {
  const span = g.base - g.top

  // 10 dB rungs when they fit, else 20, so labels don't collide.
  const step = span / (SPECTRUM_FLOOR_DB / 10) >= g.fs * 1.9 ? 10 : 20

  ctx.font = `${g.fs}px ${view.font}`
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'right'

  for (let db = 0; db >= -SPECTRUM_FLOOR_DB; db -= step) {
    const y = Math.round(g.base - (1 + db / SPECTRUM_FLOOR_DB) * span) + 0.5

    // 0 dB and -60 dB are the frame edges; don't redraw them as gridlines.
    if (db < 0 && db > -SPECTRUM_FLOOR_DB) {
      ctx.strokeStyle = cssColor(c.wellLine)
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(g.padL, y)
      ctx.lineTo(view.width - g.padX, y)
      ctx.stroke()
    }

    ctx.fillStyle = cssColor(c.ink3)
    ctx.fillText(`${db}`, g.padL - 4, y)
  }
}

export function useSpectrumRender(): CanvasOptions {
  const theme = useTheme()
  const { reedErrors, scalePitch } = useTuner()

  function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view)
    const c = theme.current.value.colors

    const accent = cssColor(c.accent)
    const green = cssColor(c.goal)
    const wellLine = cssColor(c.wellLine)

    // Nothing tracked: draw nothing, else a flat line reads as silence.
    if (!scalePitch.value || !live.hasSpectrum) return

    ctx.strokeStyle = wellLine
    ctx.lineWidth = 1
    for (let cent = -40; cent <= 40; cent += 10) {
      const x = Math.round(g.xOf(cent)) + 0.5
      ctx.beginPath()
      ctx.moveTo(x, g.top)
      ctx.lineTo(x, g.base)
      ctx.stroke()
    }

    ctx.strokeStyle = wellLine
    ctx.beginPath()
    ctx.moveTo(g.padL, Math.round(g.top) + 0.5)
    ctx.lineTo(view.width - g.padX, Math.round(g.top) + 0.5)
    ctx.stroke()

    drawDecibels(ctx, view, g, c)

    ctx.setLineDash([4, 3])
    ctx.strokeStyle = green
    ctx.lineWidth = 1.5
    for (const err of reedErrors.value) {
      const x = g.xOf(err.goal)
      ctx.beginPath()
      ctx.moveTo(x, g.top)
      ctx.lineTo(x, g.base)
      ctx.stroke()
    }
    ctx.setLineDash([])

    ctx.strokeStyle = cssColor(c.ink3)
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(Math.round(g.xOf(0)) + 0.5, g.top)
    ctx.lineTo(Math.round(g.xOf(0)) + 0.5, g.base)
    ctx.stroke()

    const path = new Path2D()
    path.moveTo(g.padL, g.base)

    for (let i = 0; i < SPECTRUM_COLUMNS; i++) {
      const mag = Math.max(0, Math.min(1, live.spectrum[i] ?? 0))
      const x = g.padL + (i / (SPECTRUM_COLUMNS - 1)) * g.inner
      path.lineTo(x, g.base - mag * (g.base - g.top))
    }

    path.lineTo(view.width - g.padX, g.base)
    path.closePath()

    const fill = ctx.createLinearGradient(0, g.top, 0, g.base)
    fill.addColorStop(0, alpha(accent, 0.34))
    fill.addColorStop(1, alpha(accent, 0.02))
    ctx.fillStyle = fill
    ctx.fill(path)

    ctx.strokeStyle = accent
    ctx.lineWidth = 1.6
    ctx.beginPath()
    for (let i = 0; i < SPECTRUM_COLUMNS; i++) {
      const mag = Math.max(0, Math.min(1, live.spectrum[i] ?? 0))
      const x = g.padL + (i / (SPECTRUM_COLUMNS - 1)) * g.inner
      const y = g.base - mag * (g.base - g.top)

      if (i === 0) ctx.moveTo(x, y)
      else ctx.lineTo(x, y)
    }
    ctx.stroke()

    const fs = g.fs

    ctx.strokeStyle = cssColor(c.neutral)
    ctx.globalAlpha = 0.35
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(g.padL, Math.round(g.base) + 0.5)
    ctx.lineTo(view.width - g.padX, Math.round(g.base) + 0.5)
    ctx.stroke()
    ctx.globalAlpha = 1

    ctx.fillStyle = cssColor(c.ink3)
    ctx.font = `${fs}px ${view.font}`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'top'

    for (let cent = -40; cent <= 40; cent += 20) {
      ctx.fillRect(Math.round(g.xOf(cent)) + 0.5, g.base + 1, 1, g.lane * 0.2)
      ctx.fillText(cent > 0 ? `+${cent}` : String(cent), g.xOf(cent), g.base + g.lane * 0.35)
    }
  }

  return { draw }
}

export { SPECTRUM_CENTS }
