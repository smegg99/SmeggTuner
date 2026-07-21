import type { Ref } from 'vue'
import { useTheme } from 'vuetify'
import { cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { useConfigSync } from '~/composables/useConfigSync'
import { NOTE_BANDS, NOTE_MIN, live, useTuner } from '~/composables/useTuner'
import { bandToFrac } from '~/utils/scale'
import { noteName } from '~/utils/tuning'
import { GUTTER, geometry, isSharp, type Geometry } from './notesGeometry'

export { bandAt } from './notesGeometry'

export function useNotesRender(hover: Ref<number>): CanvasOptions {
  const theme = useTheme()
  const { config } = useConfigSync()
  const { note, noteName: tracked } = useTuner()

  // Naming tables live in utils/tuning.ts, in step with core/tuning/notes.go.
  const nameOf = (midi: number) => noteName(midi, config.tuner?.scale_naming)

  function drawBars(ctx: CanvasRenderingContext2D, g: Geometry, band: number, c: Record<string, unknown>) {
    const ink = cssColor(c.ink)
    const ink3 = cssColor(c.ink3)

    for (let i = 0; i < NOTE_BANDS; i++) {
      // Read from the module buffer, not a Vue ref: 105 floats at 12Hz would thrash reactivity.
      const height = bandToFrac(live.bands[i] ?? 0) * (g.base - g.top)

      ctx.fillStyle = i === band ? ink : ink3
      ctx.fillRect(g.xOf(i) + g.bw * 0.12, g.base - height, Math.max(1, g.bw * 0.76), height)
    }
  }

  function drawLane(ctx: CanvasRenderingContext2D, view: CanvasView, g: Geometry, neutral: string) {
    ctx.strokeStyle = neutral
    ctx.globalAlpha = 0.35
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(g.padX, Math.round(g.base) + 0.5)
    ctx.lineTo(view.width - g.padX, Math.round(g.base) + 0.5)
    ctx.stroke()

    for (let i = 0; i < NOTE_BANDS; i++) {
      const midi = i + NOTE_MIN
      const x = Math.round(g.xOf(i) + g.bw / 2) + 0.5
      const isC = midi % 12 === 0
      const sharp = isSharp(midi)

      ctx.strokeStyle = neutral
      ctx.globalAlpha = isC ? 0.55 : sharp ? 0.16 : 0.3
      ctx.lineWidth = isC ? 1.5 : 1

      ctx.beginPath()
      ctx.moveTo(x, g.base + 1)
      ctx.lineTo(x, g.base + g.lane * (isC ? 0.4 : sharp ? 0.14 : 0.26))
      ctx.stroke()
    }

    ctx.globalAlpha = 1
  }

  function drawNames(ctx: CanvasRenderingContext2D, view: CanvasView, g: Geometry, band: number, c: Record<string, unknown>) {
    const y = g.base + g.lane * 0.68
    const label = tracked.value
    const tx = g.xOf(band) + g.bw / 2

    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    ctx.font = `700 ${g.fs}px ${view.font}`
    const trackedWidth = label ? ctx.measureText(label).width : 0

    ctx.font = `500 ${g.fs}px ${view.font}`
    ctx.fillStyle = cssColor(c.ink3)

    for (let i = 0; i < NOTE_BANDS; i++) {
      const midi = i + NOTE_MIN
      if (midi % 12) continue

      const x = g.xOf(i) + g.bw / 2
      const width = ctx.measureText(nameOf(midi)).width

      // Skip a C name the tracked name would overlap.
      if (label && Math.abs(x - tx) < (width + trackedWidth) / 2 + g.fs) continue

      // Clamp so names don't hang off the panel ends.
      ctx.fillText(
        nameOf(midi),
        Math.max(width / 2 + 2, Math.min(view.width - width / 2 - 2, x)),
        y,
      )
    }

    if (!label) return

    // Clamp inside the gutter so an end-of-keyboard name stays readable.
    ctx.font = `700 ${g.fs}px ${view.font}`
    ctx.fillStyle = cssColor(c.ink)
    ctx.fillText(
      label,
      Math.max(trackedWidth / 2 + 2, Math.min(view.width - trackedWidth / 2 - 2, tx)),
      y,
    )
  }

  function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view, view.width * GUTTER)
    const c = theme.current.value.colors
    const neutral = cssColor(c.neutral)
    const band = note.value - NOTE_MIN

    ctx.strokeStyle = cssColor(c.wellLine)
    ctx.lineWidth = 1
    for (let i = 0; i < NOTE_BANDS; i++) {
      if ((i + NOTE_MIN) % 12) continue
      const x = Math.round(g.xOf(i)) + 0.5
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, g.base)
      ctx.stroke()
    }

    if (hover.value >= 0 && hover.value !== band) {
      ctx.fillStyle = neutral
      ctx.globalAlpha = 0.1
      ctx.fillRect(g.xOf(hover.value), 0, g.bw, g.base)
      ctx.globalAlpha = 1
    }

    if (band >= 0 && band < NOTE_BANDS) {
      ctx.fillStyle = neutral
      ctx.globalAlpha = 0.14
      ctx.fillRect(g.xOf(band), 0, g.bw, g.base)
      ctx.globalAlpha = 1
    }

    drawBars(ctx, g, band, c)
    drawLane(ctx, view, g, neutral)
    drawNames(ctx, view, g, band, c)
  }

  return { draw }
}
