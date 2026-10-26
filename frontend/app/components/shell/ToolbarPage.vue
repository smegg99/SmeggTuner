<!-- app/components/shell/ToolbarPage.vue -->
<template>
  <!-- Every room but Tune is named here, not in a title strip on its card, to save the card's height. -->
  <div class="bar__page">
    <span class="bar__title">{{ title }}</span>
    <span
      v-if="pageMeta"
      class="bar__meta"
    >{{ pageMeta }}</span>
  </div>
</template>

<script setup lang="ts">
import { useShell } from '~/composables/useShell'

defineProps<{ title: string }>()
const { pageMeta } = useShell()
</script>

<style scoped>
.bar__page {
  align-items: baseline;
  display: flex;
  gap: var(--gap);
  min-width: 0;
  padding-inline-start: 4px;
}

.bar__title {
  color: rgb(var(--v-theme-ink));
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.bar__meta {
  color: rgb(var(--v-theme-ink3));
  font-size: 11.5px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
