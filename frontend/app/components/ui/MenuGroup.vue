<template>
  <div class="group">
    <span
      v-if="label"
      class="group__label"
    >{{ label }}</span>
    <slot />
  </div>
</template>

<script setup lang="ts">
// Menu rows under a caption; groups separate by a soft hairline, not a v-divider.
defineProps<{ label?: string }>()
</script>

<style scoped>
.group {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.group + .group {
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  margin-top: 3.2px;
  padding-top: 3.2px;
}

.group__label {
  color: rgb(var(--v-theme-ink3));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.06em;
  padding: 4px 9.9px 2.8px;
  text-transform: uppercase;
}
</style>
