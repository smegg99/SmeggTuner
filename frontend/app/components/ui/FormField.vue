<template>
  <div class="field">
    <label
      class="field__label"
      :for="id"
    >{{ label }}</label>

    <div class="field__control">
      <slot :id="id" />
    </div>

    <span
      v-if="hint"
      class="field__hint"
    >{{ hint }}</span>

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

.field__hint {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.4cqh;
  line-height: 1.45;
  max-width: 46ch;
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
