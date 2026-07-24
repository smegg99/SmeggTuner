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
  gap: 0.5cqh;
  min-width: 0;
}

.field__label {
  color: rgb(var(--v-theme-ink));
  font-size: 1.55cqh;
  font-weight: 600;
  line-height: 1.2;
}

/* A dotted underline is the only sign there is more to read. */
.field__label--hinted {
  cursor: help;
  text-decoration: underline dotted rgb(var(--v-theme-ink3));
  text-underline-offset: 0.35cqh;
  width: fit-content;
}

.field__control {
  min-width: 0;
}

/* warn, not red: red is reserved for a measured reed */
.field__error {
  color: rgb(var(--v-theme-warn));
  font-size: 1.4cqh;
}
</style>
