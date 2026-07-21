<template>
  <div
    class="text"
    :class="{ 'text--disabled': disabled, 'text--bad': invalid }"
  >
    <textarea
      v-if="rows"
      :id="id"
      class="text__input text__input--area"
      :rows="rows"
      :maxlength="maxlength"
      :value="modelValue"
      :disabled="disabled"
      :placeholder="placeholder"
      autocomplete="off"
      @input="onInput"
      @blur="emit('commit')"
    />

    <input
      v-else
      :id="id"
      class="text__input"
      type="text"
      autocomplete="off"
      :maxlength="maxlength"
      :value="modelValue"
      :disabled="disabled"
      :placeholder="placeholder"
      @input="onInput"
      @blur="emit('commit')"
      @keyup.enter="onEnter"
    >
  </div>
</template>

<script setup lang="ts">
// Not a v-text-field (its label/hint/density prop fight for vertical space in a row). Set rows to make it a textarea.
defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
  rows?: number
  maxlength?: number
  invalid?: boolean
  id?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [string]

  /** value settled (blur or Enter); owner may validate and persist. NOT submit: blur fires on any focus change, use enter for that. */
  'commit': []

  /** Enter only: the whole form is done, not just this field. */
  'enter': []
}>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement | HTMLTextAreaElement).value)
}

function onEnter() {
  emit('commit')
  emit('enter')
}
</script>

<style scoped>
.text {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: flex;
  min-width: 0;
  padding-inline: 0.7cqw;
}

.text:focus-within {
  border-color: rgb(var(--v-theme-ink2));
}

.text--disabled {
  opacity: 0.38;
}

/* error shown on the control, not under it, so the form doesn't jump as you fill it */
.text--bad {
  border-color: rgb(var(--v-theme-error));
}

.text__input {
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-family: inherit;
  font-size: 1.8cqh;
  height: 4.4cqh;
  min-width: 0;
  outline: none;
  width: 100%;
}

.text__input--area {
  font-size: 1.7cqh;
  height: auto;
  line-height: 1.45;
  padding-block: 1cqh;
  resize: none;
}

.text__input::placeholder {
  color: rgb(var(--v-theme-ink3));
}
</style>
