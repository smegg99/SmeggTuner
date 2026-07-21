<template>
  <UiPanelCard :title="title">
    <template #meta>
      <slot name="meta" />
    </template>

    <div
      ref="host"
      class="surface"
      :class="{ 'surface--interactive': interactive }"
      @pointerdown="onDown"
      @pointermove="emit('pointermove', $event)"
      @pointerleave="emit('pointerleave', $event)"
      @pointerup="emit('pointerup', $event)"
      @pointercancel="emit('pointercancel', $event)"
      @click="emit('click', $event)"
    >
      <canvas
        ref="canvas"
        role="img"
        :aria-label="label"
      />
    </div>
  </UiPanelCard>
</template>

<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { useTheme } from 'vuetify'
import { useCanvas } from '~/composables/useCanvas'
import type { CanvasOptions } from '~/composables/useCanvas'
import { onData } from '~/composables/useTuner'

// Canvas panel wired to the frame loop; drawing lives in a renderer.
const props = defineProps<{
  title?: string
  /** a11y label; a canvas is an image and must say what it is */
  label: string
  renderer: CanvasOptions
  /** takes the pointer */
  interactive?: boolean
}>()

const emit = defineEmits<{
  pointerdown: [PointerEvent]
  pointermove: [PointerEvent]
  pointerleave: [PointerEvent]
  pointerup: [PointerEvent]
  pointercancel: [PointerEvent]
  click: [MouseEvent]
}>()

// Capture the pointer so a release outside still delivers pointerup here, else a tone can sound forever.
function onDown(event: PointerEvent) {
  const target = event.currentTarget
  if (target instanceof HTMLElement) target.setPointerCapture(event.pointerId)

  emit('pointerdown', event)
}

const theme = useTheme()
const { host, canvas, paint, redraw } = useCanvas(props.renderer)

// Hot path: paint straight from useTuner's buffer, bypassing Vue reactivity.
const off = onData(paint)

// theme change repaints, including the cached static layer
watch(() => theme.global.name.value, redraw)

onBeforeUnmount(off)

defineExpose({ paint, redraw })
</script>

<style scoped>
.surface {
  height: 100%;
  width: 100%;
}

.surface--interactive {
  cursor: pointer;
}

.surface canvas {
  display: block;
  height: 100%;
  width: 100%;
}
</style>
