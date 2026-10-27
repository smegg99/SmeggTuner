<template>
  <UiPanel
    :title="bare ? undefined : title"
    flush
    class="tuner-panel"
  >
    <template
      v-if="$slots.head"
      #head
    >
      <slot name="head" />
    </template>
    <template
      v-if="!bare"
      #meta
    >
      <slot name="meta" />
    </template>

    <div
      ref="host"
      class="surface"
      :class="{ 'surface--interactive': interactive }"
      @pointerdown="onDown"
      @pointermove="!tracking &amp;&amp; emit('pointermove', $event)"
      @pointerleave="emit('pointerleave', $event)"
      @click="emit('click', $event)"
    >
      <canvas
        ref="canvas"
        role="img"
        :aria-label="label"
      />
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { trackPointer } from '@smegg99/s99ab/browser'
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
  /** no head strip: the picture is the whole panel */
  bare?: boolean
}>()

const emit = defineEmits<{
  pointerdown: [PointerEvent]
  pointermove: [PointerEvent]
  pointerleave: [PointerEvent]
  pointerup: [PointerEvent]
  pointercancel: [PointerEvent]
  click: [MouseEvent]
}>()

let tracking = false
let stopPointer: (() => void) | undefined
onBeforeUnmount(() => stopPointer?.())

function onDown(event: PointerEvent) {
  if (!props.interactive || event.button !== 0) return
  stopPointer?.()
  tracking = true
  stopPointer = trackPointer(event, {
    move: next => emit('pointermove', next),
    end: () => {
      tracking = false
      emit('pointerup', event)
    },
  })
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
