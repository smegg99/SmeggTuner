import { useTheme } from 'vuetify'
import { cssColor } from '~/composables/canvasColor'
import type { CanvasOptions, CanvasView } from '~/composables/useCanvas'
import { useTuner } from '~/composables/useTuner'
import { RULER_CENTS } from '~/utils/scale'
import { liveReedsUsable } from '~/utils/tuning'
import { geometry, type Geometry } from './rulerGeometry'

// Needles ease at display rate so they don't step.
const EASE = 12 // higher is snappier
const SETTLED = 0.005 // below this a needle has arrived; without it it eases forever

export function useRulerRender(): CanvasOptions {
  const theme = useTheme()
  const { reeds, reedErrors, reedsSeparated, reedsFromBeat } = useTuner()

  // Where each needle currently is, not where its reading says it should be.
  const shown: number[] = []

  const colors = () => theme.current.value.colors

  // Ruled scale and numbers never move: rasterised once then blitted.
  function background(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view)
    const c = colors()
    const neutral = cssColor(c.neutral)

    ctx.lineWidth = 1
    for (let cent = -RULER_CENTS; cent <= RULER_CENTS; cent++) {
      const x = Math.round(g.xOf(cent)) + 0.5
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

  function drawNeedle(ctx: CanvasRenderingContext2D, view: CanvasView, g: Geometry, i: number) {
    const reed = reeds.value[i]
    const cent = shown[i]
    if (!reed || cent === undefined) return

    const c = colors()
    const red = cssColor(c.reed)
    const x = g.xOf(cent)
    const head = Math.max(5, Math.min(16, view.height * 0.045))

    ctx.strokeStyle = red
    ctx.lineWidth = g.needle
    ctx.beginPath()
    ctx.moveTo(x, g.scaleTop)
    ctx.lineTo(x, g.scaleBot)
    ctx.stroke()

    // Filled head marks where the reed is; the bare goal rule is where it belongs.
    ctx.fillStyle = red
    ctx.beginPath()
    ctx.moveTo(x - head * 0.6, g.scaleTop - 1)
    ctx.lineTo(x + head * 0.6, g.scaleTop - 1)
    ctx.lineTo(x, g.scaleTop + head)
    ctx.closePath()
    ctx.fill()

    const label = reed.freq.toFixed(2)
    ctx.font = `${g.hzFs}px ${view.font}`

    const tw = ctx.measureText(label).width + g.hzFs
    const bh = g.hzFs * 1.6
    const bx = Math.max(2, Math.min(view.width - tw - 2, x - tw / 2))

    ctx.fillStyle = cssColor(c.chrome2)
    ctx.strokeStyle = red
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.rect(Math.round(bx) + 0.5, Math.round(g.hzY - bh / 2) + 0.5, tw, bh)
    ctx.fill()
    ctx.stroke()

    ctx.fillStyle = red
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(label, bx + tw / 2, g.hzY)
  }

  function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
    const g = geometry(view)

    // Unsplit peaks are not reeds: hold the bare scale rather than draw needles.
    if (!liveReedsUsable(reedsSeparated.value, reedsFromBeat.value)) {
      return
    }

    ctx.strokeStyle = cssColor(colors().goal)
    ctx.lineWidth = g.needle
    for (const err of reedErrors.value) {
      const x = g.xOf(err.goal)
      ctx.beginPath()
      ctx.moveTo(x, g.scaleTop)
      ctx.lineTo(x, g.scaleBot)
      ctx.stroke()
    }

    for (let i = 0; i < shown.length; i++) drawNeedle(ctx, view, g, i)
  }

  function advance(dt: number): boolean {
    const target = reedErrors.value

    // Drop needles for reeds that stopped sounding.
    if (shown.length !== target.length) shown.length = target.length

    // Frame-rate independent: same time at 60 or 144 Hz.
    const k = 1 - Math.exp(-EASE * dt)
    let moving = false

    for (let i = 0; i < target.length; i++) {
      const want = target[i]!.curr
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
