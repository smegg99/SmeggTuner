import { useTheme } from 'vuetify'
import { alpha, cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { live } from '~/composables/useTuner'

const GRID_COLUMNS = 8

export function useInputRender(): CanvasOptions {
  const theme = useTheme()

  function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const c = theme.current.value.colors
    const neutral = cssColor(c.neutral)
    const accent = cssColor(c.accent)

    const padX = view.width * 0.011
    const mid = view.height / 2
    const amp = view.height * 0.4

    ctx.strokeStyle = cssColor(c.wellLine)
    ctx.lineWidth = 1

    for (let i = 1; i < GRID_COLUMNS; i++) {
      const x = Math.round(padX + ((view.width - padX * 2) / GRID_COLUMNS) * i) + 0.5
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, view.height)
      ctx.stroke()
    }

    for (const f of [-0.5, 0.5]) {
      const y = Math.round(mid + f * amp) + 0.5
      ctx.beginPath()
      ctx.moveTo(padX, y)
      ctx.lineTo(view.width - padX, y)
      ctx.stroke()
    }

    const points = live.waveLength
    if (points > 1) {
      const xOf = (i: number) => padX + (i / (points - 1)) * (view.width - padX * 2)
      const yOf = (i: number) => mid - Math.max(-1, Math.min(1, live.wave[i] ?? 0)) * amp

      const path = new Path2D()
      path.moveTo(padX, mid)
      for (let i = 0; i < points; i++) path.lineTo(xOf(i), yOf(i))
      path.lineTo(view.width - padX, mid)
      path.closePath()

      const fill = ctx.createLinearGradient(0, mid - amp, 0, mid + amp)
      fill.addColorStop(0, alpha(accent, 0.3))
      fill.addColorStop(0.5, alpha(accent, 0.1))
      fill.addColorStop(1, alpha(accent, 0.3))
      ctx.fillStyle = fill
      ctx.fill(path)

      ctx.strokeStyle = accent
      ctx.lineWidth = 1.2
      ctx.beginPath()
      for (let i = 0; i < points; i++) {
        if (i === 0) ctx.moveTo(xOf(i), yOf(i))
        else ctx.lineTo(xOf(i), yOf(i))
      }
      ctx.stroke()
    }

    ctx.strokeStyle = neutral
    ctx.globalAlpha = 0.4
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(padX, Math.round(mid) + 0.5)
    ctx.lineTo(view.width - padX, Math.round(mid) + 0.5)
    ctx.stroke()
    ctx.globalAlpha = 1
  }

  return { draw }
}
