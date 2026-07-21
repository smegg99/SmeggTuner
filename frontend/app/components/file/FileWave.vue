<template>
  <div
    ref="host"
    class="wave"
    @wheel.prevent="onWheel"
    @pointerdown="onDown"
    @pointermove="onMove"
    @pointerup="onUp"
    @pointercancel="onUp"
  >
    <canvas
      ref="canvas"
      role="img"
      :aria-label="t('file.waveform', { name: state.name })"
    />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useTheme } from 'vuetify'
import { useCanvas } from '~/composables/useCanvas'
import { laneHeight, useWaveRender } from '~/composables/render/useWaveRender'
import { useTransport } from '~/composables/useTransport'

// Range clamping lives in Go (core/audio), not here; the view only draws the result.
const { t } = useI18n()
const theme = useTheme()
const { state, view, peaks, dragging, advancing, ghostAt, seek, select, zoom, pan, loadPeaks } = useTransport()

const { host, canvas, paint, redraw } = useCanvas(useWaveRender())

function timeAt(event: PointerEvent | WheelEvent): number {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const frac = (event.clientX - rect.left) / Math.max(rect.width, 1)
  return view.from + frac * (view.to - view.from)
}

function onTimeline(event: PointerEvent): boolean {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  return event.clientY - rect.top <= laneHeight(rect.height)
}

let mode: 'scrub' | 'select' | null = null
let anchor = 0

function onDown(event: PointerEvent) {
  if (!state.available) return
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)

  const at = timeAt(event)
  dragging.value = true

  if (onTimeline(event)) {
    mode = 'scrub'
    void seek(at)
    return
  }

  // A click with no drag makes an empty range, which Go reads as "all of it".
  mode = 'select'
  anchor = at
}

function onMove(event: PointerEvent) {
  if (!mode) return

  const at = timeAt(event)
  if (mode === 'scrub') {
    void seek(at)
    return
  }
  void select(Math.min(anchor, at), Math.max(anchor, at))
}

function onUp() {
  mode = null
  dragging.value = false
}

function onWheel(event: WheelEvent) {
  if (!state.available) return

  if (event.shiftKey) {
    pan((event.deltaY / 200) * (view.to - view.from) * 0.25)
    return
  }
  zoom(event.deltaY > 0 ? 1.2 : 1 / 1.2, timeAt(event))
}

// No peak cache: the Go min/max sweep costs less than the IPC call to fetch it.
async function reload() {
  const width = host.value?.clientWidth ?? 0
  if (width <= 0) return // laid out but not measured yet; the observer will be back

  await loadPeaks(width)
  paint()
}

// ResizeObserver, not just the view watch: the watch fires at width 0 before layout.
const observer = new ResizeObserver(() => void reload())

onMounted(() => {
  if (host.value) observer.observe(host.value)
  void reload()
})

onBeforeUnmount(() => observer.disconnect())

watch(() => [view.from, view.to, state.available, state.duration], reload)

// The selection is drawn over unchanged samples, not fetched.
watch(() => [state.from, state.to], paint)

// Needle repaints at frame rate, not event rate: event-driven paint stepped and lagged.
let frame = 0

function follow() {
  frame = requestAnimationFrame(follow)
  // advancing, not "not paused": a stopped needle would coast off a dead anchor.
  if (advancing.value) paint()
}

onMounted(() => {
  frame = requestAnimationFrame(follow)
})

onBeforeUnmount(() => cancelAnimationFrame(frame))

// One last paint on stop/seek/pause so the needle lands where the sound did.
watch(() => [state.position, advancing.value], paint)

// The ghost needs its own paint: a stopped transport is not repainting for the needle.
watch(ghostAt, paint)
watch(peaks, paint)
watch(() => theme.global.name.value, redraw)

// Clear a live drag on unmount, or the transport keeps ignoring incoming playhead events.
onBeforeUnmount(() => {
  dragging.value = false
})

defineExpose({ reload })
</script>

<style scoped>
.wave {
  cursor: text;
  height: 100%;
  width: 100%;
}

.wave canvas {
  display: block;
  height: 100%;
  width: 100%;
}
</style>
