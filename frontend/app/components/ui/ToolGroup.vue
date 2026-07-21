<template>
  <div
    class="group"
    role="group"
    :aria-label="label"
  >
    <UiToolKey
      v-for="item in items"
      :key="String(item.value)"
      :label="item.label"
      :icon="item.icon"
      :active="item.value === modelValue"
      :disabled="disabled"
      :title="item.title"
      @click="emit('update:modelValue', item.value)"
    />
  </div>
</template>

<script setup lang="ts">
// A segmented control: one selection always in force, never empty.
export interface ToolItem {
  value: string | number | boolean
  label?: string
  icon?: string
  title?: string
}

defineProps<{
  modelValue: string | number | boolean | undefined
  items: readonly ToolItem[]
  label?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [string | number | boolean]
}>()
</script>

<style scoped>
.group {
  display: inline-flex;
  flex: 0 0 auto;
}

/* the keys butt against each other and share their borders */
.group :deep(.key) {
  border-radius: 0;
}

.group :deep(.key:first-child) {
  border-radius: 2px 0 0 2px;
}

.group :deep(.key:last-child) {
  border-radius: 0 2px 2px 0;
}

.group :deep(.key + .key) {
  margin-inline-start: -1px;
}

/* the pressed face rides over its neighbour's border, or the seam doubles */
.group :deep(.key--active) {
  position: relative;
  z-index: 1;
}
</style>
