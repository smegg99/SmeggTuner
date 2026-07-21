<template>
  <span class="level-control">
    <UiToolKey
      :icon="on ? onIcon : offIcon"
      :active="active"
      :title="toggleTitle"
      @click="emit('toggle')"
    />
    <UiLevelSlider
      :model-value="modelValue"
      :disabled="!on"
      :title="title"
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
  gap: 0.3cqw;
}
</style>
