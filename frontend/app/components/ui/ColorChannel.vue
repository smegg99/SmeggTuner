<!-- app/components/ui/ColorChannel.vue -->
<template>
  <!-- One HSV channel of the colour picker: a labelled gradient band with a native range over it. -->
  <div class="cp__row">
    <span class="cp__name">{{ label }}</span>
    <span
      class="cp__slider"
      :style="{ background: track }"
    >
      <span
        class="cp__thumb"
        :style="{ insetInlineStart: `${(value / max) * 100}%` }"
      />
      <input
        class="cp__input"
        type="range"
        min="0"
        :max="max"
        step="1"
        :value="value"
        :aria-label="label"
        @input="emit('set', Number(($event.target as HTMLInputElement).value))"
      >
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ label: string, value: number, max: number, track: string }>()
const emit = defineEmits<{ set: [number] }>()
</script>

<style scoped>
.cp__row {
  align-items: center;
  display: flex;
  gap: var(--gap);
}

.cp__name {
  color: rgb(var(--v-theme-ink2));
  flex: 0 0 auto;
  font-size: 11px;
  inline-size: 86.5px;
}

.cp__slider {
  block-size: 17px;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  flex: 1 1 auto;
  position: relative;
}

/* mix-blend difference inverts the thumb line against any colour under it, so it reads on every band */
.cp__thumb {
  background: #fff;
  block-size: 100%;
  inline-size: 2px;
  inset-block-start: 0;
  mix-blend-mode: difference;
  pointer-events: none;
  position: absolute;
  transform: translateX(-50%);
}

.cp__input {
  block-size: 100%;
  cursor: pointer;
  inline-size: 100%;
  inset: 0;
  margin: 0;
  opacity: 0;
  position: absolute;
}

.cp__slider:focus-within {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 2px;
}
</style>
