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
  margin-top: 0.4cqh;
  padding-top: 0.4cqh;
}

.group__label {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.25cqh;
  font-weight: 600;
  letter-spacing: 0.06em;
  padding: 0.5cqh 0.9cqw 0.35cqh;
  text-transform: uppercase;
}
</style>
