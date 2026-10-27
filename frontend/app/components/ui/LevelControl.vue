<template>
  <span class="level-control">
    <UiToolKey
      :icon="on ? onIcon : offIcon"
      :active="active"
      :title="toggleTitle"
      @click="emit('toggle')"
    />
    <v-slider
      :model-value="modelValue"
      :disabled="!on"
      :title="title"
      :min="0"
      :max="1"
      :step="0.01"
      :thumb-size="10"
      hide-details
      class="tuner-volume"
      :aria-label="title"
      @update:model-value="v => emit('update:modelValue', v)"
    />
  </span>
</template>

<script setup lang="ts">
// Mute key + its slider as one control; off disables the slider (still visible) rather than hiding it.
withDefaults(defineProps<{
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
</script>

<style scoped>
.level-control {
  align-items: center;
  display: inline-flex;
  flex: 0 0 auto;
  gap: 12px;
  padding-inline-end: 6px;
}
.tuner-volume {
  width: 96px;
  flex: 0 0 96px;
  margin: 0;
}
.tuner-volume :deep(.v-input__control) { margin-inline: 0; min-height: 0; }
.tuner-volume :deep(.v-slider__container) { min-height: var(--s99ab-control-h); }
.tuner-volume :deep(.v-slider-track) { border-radius: 2px; }
</style>
