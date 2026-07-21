<template>
  <label
    class="level"
    :class="{ 'level--off': disabled }"
  >
    <span class="level__track">
      <span
        class="level__fill"
        :style="{ inlineSize: `${modelValue * 100}%` }"
      />
    </span>

    <input
      class="level__input"
      type="range"
      min="0"
      max="1"
      step="0.01"
      :value="modelValue"
      :disabled="disabled"
      :title="title"
      :aria-label="title"
      @input="onInput"
    >
  </label>
</template>

<script setup lang="ts">
// Invisible native range input over a drawn track: keeps keyboard/drag/a11y, not a v-slider.
const props = defineProps<{
  modelValue: number
  title?: string
  disabled?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [number] }>()

function onInput(event: Event) {
  if (props.disabled) return
  emit('update:modelValue', Number((event.target as HTMLInputElement).value))
}
</script>

<style scoped>
.level {
  align-items: center;
  block-size: 3.9cqh;
  display: inline-flex;
  inline-size: 6.5cqw;
  position: relative;
}

.level__track {
  background: rgb(var(--v-theme-sunk));
  block-size: 0.75cqh;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 1px;
  display: block;
  inline-size: 100%;
  overflow: hidden;
}

.level__fill {
  background: rgb(var(--v-theme-ink3));
  block-size: 100%;
  display: block;
}

.level__input {
  block-size: 100%;
  cursor: pointer;
  inline-size: 100%;
  inset: 0;
  margin: 0;
  opacity: 0;
  position: absolute;
}

.level:hover .level__fill {
  background: rgb(var(--v-theme-ink2));
}

.level__input:focus-visible {
  opacity: 1;
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 1px;
}

.level--off {
  opacity: 0.38;
}

.level--off .level__input {
  cursor: default;
}
</style>
