import { useTheme } from 'vuetify'
import { cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { useDeviationUnit } from '~/composables/useDeviationUnit'
import { useTuner } from '~/composables/useTuner'
import { MISSING, shownDeviation } from '~/utils/deviation'
import { RULER_CENTS } from '~/utils/scale'
import { geometry, type Geometry } from './rulerGeometry'

// Needles ease at display rate so they don't step.
const EASE = 12 // higher is snappier
// The needle's head, as the audio preview's playhead draws it.
const HEAD = 5.5
const SETTLED = 0.005 // below this a needle has arrived; without it it eases forever

export function useRulerRender(): CanvasOptions {
  const theme = useTheme()
  const { needles, needlesDimmed, panels } = useTuner()
  const unit = useDeviationUnit()

  // Where each needle currently is, not where its reading says it should be; undefined for an invalid slot.
  const shown: (number | undefined)[] = []

  const colors = () => theme.current.value.colors

  // Ruled scale and numbers never move: rasterised once then blitted.
  function background(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view)
    const c = colors()
    const neutral = cssColor(c.neutral)

    ctx.lineWidth = 1
    for (let cent = -RULER_CENTS; cent <= RULER_CENTS; cent++) {
      const x = Math.min(view.width - 0.5, Math.round(g.xOf(cent)) + 0.5)
      const ten = cent % 10 === 0
      const five = cent % 5 === 0

      ctx.strokeStyle = neutral
      ctx.globalAlpha = ten ? 0.42 : five ? 0.24 : 0.14

      const bottom = ten
        ? g.scaleBot
        : five
          ? g.scaleTop + (g.scaleBot - g.scaleTop) * 0.42
          : g.scaleTop + (g.scaleBot - g.scaleTop) * 0.12

      ctx.beginPath()
      ctx.moveTo(x, g.scaleTop)
      ctx.lineTo(x, bottom)
      ctx.stroke()
    }
    ctx.globalAlpha = 1

    ctx.font = `600 ${g.fs}px ${view.font}`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    for (let cent = -RULER_CENTS; cent <= RULER_CENTS; cent += 10) {
      ctx.fillStyle = cent === 0 ? cssColor(c.ink) : cssColor(c.ink2)
      ctx.fillText(String(Math.abs(cent)), g.xOf(cent), g.labelY)
    }

    ctx.strokeStyle = neutral
    ctx.globalAlpha = 0.3
    ctx.beginPath()
    ctx.moveTo(g.padX, Math.round(g.scaleTop) + 0.5)
    ctx.lineTo(view.width - g.padX, Math.round(g.scaleTop) + 0.5)
    ctx.stroke()
    ctx.globalAlpha = 1

    ctx.strokeStyle = cssColor(c.ink3)
    ctx.lineWidth = 1.5
    ctx.beginPath()
    ctx.moveTo(Math.round(g.xOf(0)) + 0.5, g.scaleTop)
    ctx.lineTo(Math.round(g.xOf(0)) + 0.5, g.scaleBot)
    ctx.stroke()
  }

  // The audio preview's playhead: a 1 px line under a small head, so the scale stays readable round it.
  function drawNeedle(ctx: CanvasRenderingContext2D, g: Geometry, i: number) {
    const needle = needles.value[i]
    const cent = shown[i]
    if (!needle?.valid || cent === undefined) return

    const c = colors()
    // The reference keeps the needles with no note and dims them.
    ctx.fillStyle = cssColor(needlesDimmed.value ? c.ink3 : c.reed)
    const x = Math.round(g.xOf(cent))
    ctx.fillRect(x, g.scaleTop, 1, g.scaleBot - g.scaleTop)
    ctx.beginPath()
    ctx.moveTo(x - HEAD, g.scaleTop)
    ctx.lineTo(x + HEAD + 1, g.scaleTop)
    ctx.lineTo(x + 0.5, g.scaleTop + HEAD * 1.7)
    ctx.closePath()
    ctx.fill()
  }

  function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view)

    // Reed readouts sit at the even panels; each carries its slot's target, or Missing.
    for (let i = 0; i < panels.value.length; i += 2) {
      const goal = panels.value[i]!.target
      if (goal <= MISSING) continue
      // The bare goal line is where a reed belongs; the needle's head marks where it is.
      ctx.fillStyle = cssColor(colors().goal)
      ctx.fillRect(Math.round(g.xOf(shownDeviation(goal, unit.value))), g.scaleTop, 1, g.scaleBot - g.scaleTop)
    }

    for (let i = 0; i < shown.length; i++) drawNeedle(ctx, g, i)
  }

  function advance(dt: number): boolean {
    const target = needles.value

    if (shown.length !== target.length) shown.length = target.length

    // Frame-rate independent: same time at 60 or 144 Hz.
    const k = 1 - Math.exp(-EASE * dt)
    let moving = false

    for (let i = 0; i < target.length; i++) {
      // An invalid slot has no needle; when it turns valid again it starts at its reading.
      if (!target[i]!.valid) {
        shown[i] = undefined
        continue
      }
      const want = shownDeviation(target[i]!.dev, unit.value)
      const have = shown[i]

      if (have === undefined) {
        // A new needle starts at its reading, not sweeping from elsewhere.
        shown[i] = want
        continue
      }

      const next = have + (want - have) * k
      shown[i] = Math.abs(want - next) < SETTLED ? want : next
      if (shown[i] !== want) moving = true
    }

    return moving
  }

  return { draw, advance, background }
}
