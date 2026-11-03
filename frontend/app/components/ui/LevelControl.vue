<template>
  <span class="level-control">
    <UiToolKey
      :icon="on ? onIcon : offIcon"
      :active="active"
      :title="toggleTitle"
      @click="emit('toggle')"
    />
    <!-- Untitled, so the window's tooltip leaves it alone: its own tip says the level, and stays up
         while the slider is dragged. -->
    <span
      class="level-control__slider"
      @pointerenter="enter"
      @pointerleave="leave"
    >
      <v-slider
        :model-value="modelValue"
        :disabled="!on"
        :min="0"
        :max="1"
        :step="0.01"
        :thumb-size="10"
        hide-details
        class="tuner-volume"
        :aria-label="title"
        @update:model-value="v => emit('update:modelValue', v)"
        @start="dragging = true"
        @end="dragging = false"
      />
      <v-tooltip
        :model-value="hovering || dragging"
        activator="parent"
        :text="tip"
        :open-on-hover="false"
        :open-on-focus="false"
        :open-on-click="false"
        :transition="false"
        location="bottom"
      />
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'

// Mute key + its slider as one control; off disables the slider (still visible) rather than hiding it.
const props = withDefaults(defineProps<{
  /** the level, 0..1 */
  modelValue: number
  on: boolean
  title: string
  toggleTitle: string
  onIcon?: string
  offIcon?: string
  active?: boolean
}>(), {
  onIcon: 'mdi-volume-high',
  offIcon: 'mdi-volume-off',
  active: false,
})

const emit = defineEmits<{
  'update:modelValue': [number]
  'toggle': []
}>()

// The window's tooltip waits as long before showing.
const DELAY_MS = 450

const hovering = ref(false)
const dragging = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

const tip = computed(() => `${props.title}: ${Math.round(props.modelValue * 100)} %`)

function enter() {
  clearTimeout(timer)
  timer = setTimeout(() => {
    hovering.value = true
  }, DELAY_MS)
}

function leave() {
  clearTimeout(timer)
  hovering.value = false
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<style scoped>
.level-control {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: 12px;
  padding-inline-end: 6px;
}
.level-control__slider { display: flex; }
.tuner-volume {
  width: 96px;
  flex: 0 0 96px;
  margin: 0;
}
.tuner-volume :deep(.v-input__control) { margin-inline: 0; min-height: 0; }
.tuner-volume :deep(.v-slider__container) { min-height: var(--s99ab-control-h); }
.tuner-volume :deep(.v-slider-track) { border-radius: 2px; }
</style>
