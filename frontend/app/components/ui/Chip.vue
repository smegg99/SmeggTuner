<template>
  <span :class="['chip', { 'chip--mono': mono }]">
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  mono?: boolean
}>(), { mono: false })
</script>

<style scoped>
.chip {
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink2));
  flex: 0 0 auto;
  font-size: 11.6px;
  padding: 1.6px 5px;
}

.chip--mono {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}
</style>
