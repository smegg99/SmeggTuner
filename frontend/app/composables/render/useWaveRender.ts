import { useTheme } from 'vuetify'
import { alphaColor, cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { useTransport } from '~/composables/useTransport'
import { timeTicks } from '~/utils/time'

// Drag in the lane to scrub, below it to select.

// Exported: the pointer must agree with the picture on the scrub/select boundary.
export function laneHeight(height: number): number {
  return Math.max(12, Math.min(26, height * 0.19))
}

export function useWaveRender(): CanvasOptions {
  const theme = useTheme()
  const { state, view, peaks, livePosition, ghostAt } = useTransport()

  const xOf = (t: number, w: number) => {
    const span = view.to - view.from
    return span > 0 ? ((t - view.from) / span) * w : 0
  }

  function drawLane(ctx: CanvasRenderingContext2D, v: CanvasView, c: Record<string, unknown>, lane: number) {
    ctx.fillStyle = cssColor(c.wellLine)
    ctx.fillRect(0, 0, v.width, lane)

    const fs = Math.max(8, Math.min(13, lane * 0.52))
    ctx.font = `${fs}px ${v.font}`
    ctx.textBaseline = 'middle'
    ctx.textAlign = 'left'
    ctx.fillStyle = cssColor(c.ink3)

    for (const tick of timeTicks(view.from, view.to, v.width)) {
      const x = Math.round(xOf(tick.at, v.width))

      ctx.fillRect(x, tick.major ? lane - 5 : lane - 3, 1, tick.major ? 5 : 3)
      if (!tick.major) continue

      // Skip a label that would hang off the panel.
      if (x + ctx.measureText(tick.label).width + 6 > v.width) continue
      ctx.fillText(tick.label, x + 3, lane / 2 - 1)

      ctx.fillStyle = cssColor(c.wellLine)
      ctx.fillRect(x, lane, 1, v.height - lane)
      ctx.fillStyle = cssColor(c.ink3)
    }
  }

  function drawTrace(ctx: CanvasRenderingContext2D, v: CanvasView, c: Record<string, unknown>, lane: number) {
    const cols = peaks.value
    const mid = lane + (v.height - lane) / 2
    const half = (v.height - lane) / 2 - 3

    ctx.fillStyle = cssColor(c.wellLine)
    ctx.fillRect(0, Math.round(mid), v.width, 1)

    if (cols.length === 0 || half <= 0) return

    const g = ctx.createLinearGradient(0, lane, 0, v.height)
    g.addColorStop(0, alphaColor(c.neutral, 0.55))
    g.addColorStop(0.5, alphaColor(c.neutral, 1))
    g.addColorStop(1, alphaColor(c.neutral, 0.55))
    ctx.fillStyle = g

    const bw = v.width / cols.length
    for (let i = 0; i < cols.length; i++) {
      const p = cols[i]
      if (!p) continue

      const hi = mid - (p.max ?? 0) * half
      const lo = mid - (p.min ?? 0) * half

      // Floor width/height at 1px so a quiet column doesn't vanish as a gap.
      ctx.fillRect(i * bw, hi, Math.max(bw, 1), Math.max(lo - hi, 1))
    }
  }

  function drawSelection(ctx: CanvasRenderingContext2D, v: CanvasView, c: Record<string, unknown>, lane: number) {
    // Whole file selected is no selection: nothing to light.
    if (state.from <= 0 && state.to >= state.duration) return

    const a = xOf(state.from, v.width)
    const b = xOf(state.to, v.width)

    ctx.fillStyle = alphaColor(c.ink, 0.07)
    ctx.fillRect(a, lane, b - a, v.height - lane)

    ctx.fillStyle = cssColor(c.ink3)
    for (const x of [a, b]) {
      ctx.fillRect(Math.round(x), lane, 1, v.height - lane)
    }
  }

  // The ghost marks where the latched reading was measured; ghostAt decides if it's shown.
  function drawGhost(ctx: CanvasRenderingContext2D, v: CanvasView, c: Record<string, unknown>, lane: number) {
    const at = ghostAt.value
    if (at === null) return

    const x = Math.round(xOf(at, v.width))
    if (x < 0 || x > v.width) return

    ctx.fillStyle = alphaColor(c.ink3, 0.8)

    // Dashed so it isn't mistaken for a second playhead.
    for (let y = lane; y < v.height; y += 6) {
      ctx.fillRect(x, y, 1, 3)
    }
  }

  function drawNeedle(ctx: CanvasRenderingContext2D, v: CanvasView, c: Record<string, unknown>, lane: number) {
    // livePosition, not state.position: the ~30Hz event lags behind the measurement stream. See useTransport.
    const x = Math.round(xOf(livePosition(), v.width))
    if (x < -8 || x > v.width + 8) return

    const red = cssColor(c.reed)

    ctx.fillStyle = red
    ctx.fillRect(x, 0, 1, v.height)

    const g = Math.max(4, lane * 0.34)
    ctx.beginPath()
    ctx.moveTo(x - g, 0)
    ctx.lineTo(x + g + 1, 0)
    ctx.lineTo(x + 0.5, g * 1.7)
    ctx.closePath()
    ctx.fill()
  }

  function draw(ctx: CanvasRenderingContext2D, v: CanvasView) {
    const c = theme.current.value.colors

    ctx.fillStyle = cssColor(c.well)
    ctx.fillRect(0, 0, v.width, v.height)

    if (!state.available || view.to <= view.from) return

    const lane = laneHeight(v.height)

    drawLane(ctx, v, c, lane)
    drawTrace(ctx, v, c, lane)
    drawSelection(ctx, v, c, lane)
    drawGhost(ctx, v, c, lane)
    drawNeedle(ctx, v, c, lane)
  }

  return { draw }
}
