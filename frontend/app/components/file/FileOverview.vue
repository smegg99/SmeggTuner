<template>
  <div
    ref="strip"
    class="overview"
    @pointerdown="onDown"
    @pointermove="onMove"
    @pointerup="onUp"
    @pointercancel="onUp"
  >
    <div
      v-if="hasSelection"
      class="overview__selection"
      :style="{ left: pct(state.from), width: pct(state.to - state.from) }"
    />

    <div
      class="overview__window"
      :style="{ left: pct(view.from), width: pct(view.to - view.from) }"
    />

    <!-- where the reading was measured, once the sound stopped -->
    <div
      v-if="ghostAt !== null"
      class="overview__ghost"
      :style="{ left: pct(ghostAt) }"
    />

    <div
      class="overview__needle"
      :style="{ left: pct(needle) }"
    />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useTransport } from '~/composables/useTransport'

// Plain divs, not a canvas: rectangles the compositor moves without this app redrawing.
const { state, view, hasSelection, advancing, livePosition, ghostAt, scrollTo } = useTransport()

// Same playhead as the waveform (livePosition), so the two needles never disagree.
const needle = ref(0)
let frame = 0

function follow() {
  frame = requestAnimationFrame(follow)
  needle.value = livePosition()
}

onMounted(() => {
  frame = requestAnimationFrame(follow)
})

onBeforeUnmount(() => cancelAnimationFrame(frame))

// nothing to animate while stopped, but the needle must still land where a seek/stop put it
watch(() => [state.position, advancing.value], () => {
  needle.value = livePosition()
})

const strip = ref<HTMLElement>()
let held = false

const pct = (seconds: number) => `${(seconds / Math.max(state.duration, 1e-9)) * 100}%`

function centreOn(event: PointerEvent) {
  const rect = strip.value?.getBoundingClientRect()
  if (!rect) return

  const at = ((event.clientX - rect.left) / Math.max(rect.width, 1)) * state.duration
  scrollTo(at - (view.to - view.from) / 2)
}

function onDown(event: PointerEvent) {
  if (!state.available) return
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
  held = true
  centreOn(event)
}

function onMove(event: PointerEvent) {
  if (held) centreOn(event)
}

function onUp() {
  held = false
}
</script>

<style scoped>
.overview {
  background: rgb(var(--v-theme-sunk));
  border-top: 1px solid rgb(var(--v-theme-line));
  cursor: pointer;
  height: 100%;
  position: relative;
  width: 100%;
}

.overview__selection {
  background: rgba(var(--v-theme-neutral), 0.16);
  bottom: 0;
  position: absolute;
  top: 0;
}

.overview__window {
  border: 1px solid rgb(var(--v-theme-ink3));
  bottom: 0;
  min-width: 1px;
  position: absolute;
  top: 0;
}

.overview__needle {
  background: rgb(var(--v-theme-reed));
  bottom: 0;
  position: absolute;
  top: 0;
  width: 1px;
}

.overview__ghost {
  background: rgba(var(--v-theme-ink3), 0.75);
  bottom: 0;
  position: absolute;
  top: 0;
  width: 1px;
}
</style>
