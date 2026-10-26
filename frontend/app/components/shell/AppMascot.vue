<!-- frontend/app/components/shell/AppMascot.vue -->
<template>
  <!-- Trusted SVG from the repository's asset master. -->
  <!-- eslint-disable vue/no-v-html -->
  <span
    class="mascot"
    role="img"
    :aria-label="label"
    v-html="drawing"
  />
</template>

<script setup lang="ts">
import drawing from '../../../../assets/SmeggTunerMascot.svg?raw'

defineProps<{ label?: string }>()
</script>

<style scoped>
.mascot {
  display: inline-flex;
  color: rgb(var(--v-theme-on-surface));
}

.mascot :deep(svg) {
  display: block;
  height: var(--app-mascot-h, 160px);
  width: auto;
}

/* The master is one white outline path; it follows the theme ink instead. */
.mascot :deep(path) {
  fill: currentColor !important;
}
</style>
