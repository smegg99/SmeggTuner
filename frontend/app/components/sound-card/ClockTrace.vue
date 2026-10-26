<template>
  <!-- The clock offset over the run, drawn as the spectrum's curve is; empty until a block is measured. -->
  <div
    ref="host"
    class="trace"
  >
    <canvas
      ref="canvas"
      role="img"
      :aria-label="t('settings.calibration.screen.stat.offset')"
    />
    <span
      v-if="points.length < 2"
      class="trace__empty"
    >{{ t('settings.calibration.screen.traceEmpty') }}</span>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useTheme } from 'vuetify'
import { alpha, cssColor } from '~/composables/canvasColor'
import { useCanvas, type CanvasView } from '~/composables/useCanvas'
import type { TracePoint } from '~/composables/useSoundCard'

const props = defineProps<{ points: TracePoint[] }>()

// The narrowest offset range drawn, in percent (20 ppm), so a steady clock's jitter stays flat.
const MIN_SPAN = 0.002
// The shortest time axis, in seconds, so the first points do not fill the width.
const MIN_SECONDS = 20

const { t } = useI18n()
const theme = useTheme()

function draw(ctx: CanvasRenderingContext2D, view: CanvasView) {
  const pts = props.points
  if (pts.length < 2) return
  const accent = cssColor(theme.current.value.colors.accent)

  const t0 = pts[0]!.at
  const t1 = Math.max(pts.at(-1)!.at, t0 + MIN_SECONDS)
  let lo = Math.min(...pts.map(p => p.percent))
  let hi = Math.max(...pts.map(p => p.percent))
  if (hi - lo < MIN_SPAN) {
    const mid = (hi + lo) / 2
    lo = mid - MIN_SPAN / 2
    hi = mid + MIN_SPAN / 2
  }
  const pad = 4
  const xOf = (at: number) => pad + ((at - t0) / (t1 - t0)) * (view.width - pad * 2)
  const yOf = (p: number) => pad + (1 - (p - lo) / (hi - lo)) * (view.height - pad * 2)

  const line = new Path2D()
  pts.forEach((p, i) => (i ? line.lineTo(xOf(p.at), yOf(p.percent)) : line.moveTo(xOf(p.at), yOf(p.percent))))

  const area = new Path2D(line)
  area.lineTo(xOf(pts.at(-1)!.at), view.height)
  area.lineTo(xOf(t0), view.height)
  area.closePath()
  const fill = ctx.createLinearGradient(0, 0, 0, view.height)
  fill.addColorStop(0, alpha(accent, 0.34))
  fill.addColorStop(1, alpha(accent, 0.02))
  ctx.fillStyle = fill
  ctx.fill(area)

  ctx.strokeStyle = accent
  ctx.lineWidth = 1.6
  ctx.lineJoin = 'round'
  ctx.stroke(line)

  const last = pts.at(-1)!
  ctx.fillStyle = accent
  ctx.beginPath()
  ctx.arc(xOf(last.at), yOf(last.percent), 2.5, 0, Math.PI * 2)
  ctx.fill()
}

const { host, canvas, paint, redraw } = useCanvas({ draw })
watch(() => props.points, paint)
watch(() => theme.current.value.colors.accent, redraw)
</script>

<style scoped>
.trace {
  min-height: 0;
  position: relative;
}

.trace canvas {
  display: block;
  height: 100%;
  inset: 0;
  position: absolute;
  width: 100%;
}

.trace__empty {
  align-items: center;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 12px;
  inset: 0;
  justify-content: center;
  position: absolute;
}
</style>
