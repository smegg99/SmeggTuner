<template>
  <div class="field">
    <!-- The hint hovers on the label rather than taking a paragraph under every input. -->
    <label
      class="field__label"
      :class="{ 'field__label--hinted': hint }"
      :for="id"
    >{{ label }}<v-tooltip
      v-if="hint"
      activator="parent"
      location="top start"
      max-width="340"
      open-delay="300"
    >{{ hint }}</v-tooltip></label>

    <div class="field__control">
      <slot :id="id" />
    </div>

    <span
      v-if="error"
      class="field__error"
    >{{ error }}</span>
  </div>
</template>

<script setup lang="ts">
import { useId } from 'vue'

defineProps<{
  label: string
  hint?: string
  error?: string
}>()

const id = useId()
</script>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.field__label {
  color: rgb(var(--v-theme-ink));
  font-size: 12.4px;
  font-weight: 600;
  line-height: 1.2;
}

/* The help cursor is the whole affordance, as it is on a hinted UiFieldGroup title. A label is
   not a link and a dotted rule under every second one turns a form into a ransom note. */
.field__label--hinted {
  cursor: help;
  width: fit-content;
}

.field__control {
  min-width: 0;
}

/* warn, not red: red is reserved for a measured reed */
.field__error {
  color: rgb(var(--v-theme-warn));
  font-size: 11.2px;
}
</style>
