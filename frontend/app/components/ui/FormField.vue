<template>
  <div
    class="field"
    :class="{ 'field--row': row }"
  >
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
  /** a short control: a compact window puts it on its label's row, at the far end */
  row?: boolean
}>()

const id = useId()
</script>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--gap-tight);
  min-width: 0;
}

.field__label {
  color: rgb(var(--v-theme-ink));
  font-size: 11px;
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
  font-size: 11px;
}

/* A compact window is read at arm's length and has width to spare beside a short control: the label is
   a size larger, and a row field puts its control at the far end of the label's row. A control too
   wide for that takes the next line, still at the far end. */
@media (max-width: 959px), (max-height: 559px) {
  .field__label { font-size: 13px; }
  .field--row {
    flex-flow: row wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--gap-tight) 12px;
  }
  .field--row > .field__control { margin-inline-start: auto; }
  .field--row > .field__error { flex: 1 0 100%; }
}
</style>
