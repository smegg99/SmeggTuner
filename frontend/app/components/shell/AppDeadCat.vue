<!-- app/components/shell/AppDeadCat.vue -->
<template>
  <!-- Trusted SVG from the repository's asset master. -->
  <!-- eslint-disable vue/no-v-html -->
  <span
    class="dead"
    role="img"
    :aria-label="label"
    v-html="drawing"
  />
</template>

<script setup lang="ts">
import drawing from '../../../../assets/cat-dead.svg?raw'

defineProps<{ label?: string }>()
</script>

<style scoped>
.dead {
  display: inline-flex;
  color: rgb(var(--v-theme-on-surface));
}

.dead :deep(svg) {
  display: block;
  height: var(--app-dead-cat-h, 96px);
  width: auto;
}

/* The head is white with the face cut out of it; it follows the theme ink, so the cut-outs show the ground. */
.dead :deep(#path12) {
  fill: currentColor !important;
}
</style>
