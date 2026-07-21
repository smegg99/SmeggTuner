import { onBeforeUnmount, onMounted, ref } from 'vue'
import { addPainter, removePainter, wakeFrameLoop } from '~/composables/useFrameLoop'
import type { Painter } from '~/composables/useFrameLoop'

// Paints only when dirty: a readout (no advance) draws once and the loop stops; a needle (advance) keeps easing until settled.
// Static parts go in background(), rasterised offscreen and blitted.

// Two device pixels per CSS pixel is the end of what an eye finds on a needle, and
// fill rate costs its square; three buys nothing and WebKitGTK is where we'd pay.
const MAX_DPR = 2

export interface CanvasView {
  /** CSS pixels: the device pixel ratio is already in the transform. */
  width: number
  height: number
  /** The page's font family. A canvas cannot inherit one and has to be told. */
  font: string
}

export interface CanvasOptions {
  draw: (ctx: CanvasRenderingContext2D, view: CanvasView) => void
  /** Eases display state toward the last reading and reports whether anything
   *  moved. Omit for a readout, which has no ballistics. */
  advance?: (dt: number) => boolean
  /** The part of the picture that never moves. Redrawn on resize and redraw(). */
  background?: (ctx: CanvasRenderingContext2D, view: CanvasView) => void
}

export function useCanvas(options: CanvasOptions) {
  const host = ref<HTMLDivElement | null>(null)
  const canvas = ref<HTMLCanvasElement | null>(null)

  const view: CanvasView = { width: 0, height: 0, font: 'sans-serif' }

  let ctx: CanvasRenderingContext2D | null = null
  let layer: HTMLCanvasElement | null = null
  let layerCtx: CanvasRenderingContext2D | null = null
  let observer: ResizeObserver | null = null
  let dpr = 1
  let dirty = true
  let layerDirty = true
  let sizeDirty = false

  /** paint asks for one more frame: the reading changed. */
  function paint() {
    dirty = true
    wakeFrameLoop()
  }

  /** invalidate also throws the static layer away: the picture's size changed. */
  function invalidate() {
    layerDirty = true
    paint()
  }

  // redraw = invalidate + reread the font, for a theme/language change. The font is
  // read only here, never on the resize path: getComputedStyle forces a layout flush
  // and doing that per resize callback per canvas thrashes inside a window drag.
  function redraw() {
    readFont()
    invalidate()
  }

  function readFont() {
    const element = canvas.value
    if (element) view.font = getComputedStyle(element).fontFamily || view.font
  }

  function ratio() {
    return Math.min(MAX_DPR, window.devicePixelRatio || 1)
  }

  // Only mark the box dirty; the measure happens on the next frame (see painter).
  function onResize() {
    sizeDirty = true
    wakeFrameLoop()
  }

  function measure() {
    sizeDirty = false

    const element = canvas.value
    const box = host.value
    if (!element || !box) return

    const width = box.clientWidth
    const height = box.clientHeight
    if (width <= 0 || height <= 0) return

    const nextDpr = ratio()
    const bw = Math.round(width * nextDpr)
    const bh = Math.round(height * nextDpr)

    // Nothing actually moved: a ResizeObserver fires for layout passes that left the
    // box where it was, and assigning width reallocates/clears the backing store and
    // invalidate() throws away an expensive background. Bail if the size is unchanged.
    if (element.width === bw && element.height === bh && nextDpr === dpr) return

    dpr = nextDpr
    element.width = bw
    element.height = bh

    view.width = width
    view.height = height

    invalidate()
  }

  function drawLayer() {
    if (!options.background) return null
    if (!layer) {
      layer = document.createElement('canvas')
      layerCtx = layer.getContext('2d')
    }
    if (!layerCtx) return null

    const bw = Math.round(view.width * dpr)
    const bh = Math.round(view.height * dpr)
    if (layer.width !== bw) layer.width = bw
    if (layer.height !== bh) layer.height = bh

    layerCtx.setTransform(dpr, 0, 0, dpr, 0, 0)
    layerCtx.clearRect(0, 0, view.width, view.height)
    options.background(layerCtx, view)
    layerDirty = false
    return layer
  }

  function render() {
    if (!ctx || view.width <= 0 || view.height <= 0) return

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, view.width, view.height)

    if (options.background) {
      const cached = layerDirty ? drawLayer() : layer
      if (cached) {
        ctx.setTransform(1, 0, 0, 1, 0, 0)
        ctx.drawImage(cached, 0, 0)
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
      }
    }

    options.draw(ctx, view)
    dirty = false
  }

  const painter: Painter = {
    tick(dt) {
      // Resize is applied here, on the frame, not in the observer: a window-edge drag
      // fires the observer many times between paints, and each measure reallocates the
      // backing store and drops the background layer. Coalesced here it costs one
      // measure per frame. The pixel ratio is checked too, since moving to a screen
      // with another ratio changes no CSS size and the observer stays quiet.
      if (sizeDirty || ratio() !== dpr) measure()

      const moved = options.advance?.(dt) ?? false
      if (moved || dirty) render()
      // Only a still-easing canvas keeps the loop alive; a readout lets it stop.
      return moved
    },
  }

  onMounted(() => {
    const element = canvas.value
    const box = host.value
    if (!element || !box) return

    ctx = element.getContext('2d')
    if (!ctx) return

    readFont()
    observer = new ResizeObserver(onResize)
    observer.observe(box)
    measure()
    addPainter(painter)
  })

  onBeforeUnmount(() => {
    removePainter(painter)
    observer?.disconnect()
    observer = null
    ctx = null
    layer = null
    layerCtx = null
  })

  return { host, canvas, paint, redraw }
}
