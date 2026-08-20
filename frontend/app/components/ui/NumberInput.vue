<template>
  <div
    class="num"
    :class="{ 'num--disabled': disabled }"
  >
    <input
      :id="id"
      class="num__input"
      type="text"
      inputmode="decimal"
      autocomplete="off"
      :value="modelValue"
      :disabled="disabled"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @blur="emit('commit')"
      @keyup.enter="emit('commit')"
    >

    <span
      v-if="suffix"
      class="num__suffix"
    >{{ suffix }}</span>
  </div>
</template>

<script setup lang="ts">
// type="text"+inputmode="decimal" avoids WebKit's inconsistent spinner arrows; owner validates on commit.
defineProps<{
  modelValue: string
  suffix?: string
  disabled?: boolean
  id?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [string]
  /** blur or enter: the owner may now validate, clamp and persist */
  'commit': []
}>()
</script>

<style scoped>
.num {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: flex;
  gap: 5.5px;
  min-width: 0;
  padding-inline: 7.7px;
}

.num:focus-within {
  border-color: rgb(var(--v-theme-ink2));
}

.num--disabled {
  opacity: 0.38;
}

.num__input {
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-family: var(--font-mono);
  font-size: 14.4px;
  font-variant-numeric: tabular-nums;
  height: 35.2px;
  min-width: 0;
  outline: none;
  width: 100%;
}

.num__suffix {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-size: 12.4px;
}
</style>
