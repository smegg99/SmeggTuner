<template>
  <SmeggTunerLogo
    class="logo"
    role="img"
    :aria-label="title ?? 'SmeggTuner'"
  />
</template>

<script setup lang="ts">
import SmeggTunerLogo from '~/assets/smeggtuner-logo.svg'

// SVG greys are remapped to theme ink vars, so one file flips with the theme (via vite-svg-loader).
defineProps<{ title?: string }>()
</script>

<style scoped>
/* Height is a CSS var so the parent bar can size it to the row. */
.logo {
  display: block;
  height: var(--app-logo-h, 2cqh);
  width: auto;
}
</style>
