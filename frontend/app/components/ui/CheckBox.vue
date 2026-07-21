<template>
  <label
    class="check"
    :class="{ 'check--disabled': disabled }"
  >
    <input
      class="check__input"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      @change="emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
    >

    <!-- Custom-drawn box; a hidden native checkbox underneath keeps keyboard/focus/label/a11y. -->
    <span
      class="check__box"
      aria-hidden="true"
    >
      <svg
        class="check__tick"
        viewBox="0 0 16 16"
        fill="none"
      >
        <path
          d="M3.5 8.5 L6.5 11.5 L12.5 4.5"
          stroke="currentColor"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </span>

    <span class="check__text">
      <span class="check__label">{{ label }}</span>
      <span
        v-if="hint"
        class="check__hint"
      >{{ hint }}</span>
    </span>
  </label>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean
  label: string
  hint?: string
  disabled?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [boolean] }>()
</script>

<style scoped>
.check {
  align-items: flex-start;
  cursor: pointer;
  display: flex;
  gap: 0.7cqw;
  min-width: 0;
}

.check--disabled {
  cursor: default;
  opacity: 0.38;
}

.check__input {
  height: 0;
  opacity: 0;
  position: absolute;
  width: 0;
}

.check__box {
  align-items: center;
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 2.2cqh;
  justify-content: center;
  /* it sits on the first line of the label beside it, not on the top of the block */
  margin-block-start: 0.1cqh;
  width: 2.2cqh;
}

.check:hover .check__box {
  border-color: rgb(var(--v-theme-ink2));
}

.check__input:focus-visible + .check__box {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 1px;
}

.check__input:checked + .check__box {
  background: rgb(var(--v-theme-ink2));
  border-color: rgb(var(--v-theme-ink2));
  color: rgb(var(--v-theme-bg));
}

.check__tick {
  height: 80%;
  opacity: 0;
  width: 80%;
}

.check__input:checked + .check__box .check__tick {
  opacity: 1;
}

.check__text {
  display: flex;
  flex-direction: column;
  gap: 0.15cqh;
  min-width: 0;
}

.check__label {
  color: rgb(var(--v-theme-ink));
  font-size: 1.6cqh;
}

/* ink2, not ink3: ink3 is barely 3:1 on a well, too faint to read */
.check__hint {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.4cqh;
  line-height: 1.45;
  max-width: 46ch;
}
</style>
