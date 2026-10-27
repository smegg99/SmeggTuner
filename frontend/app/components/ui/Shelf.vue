<template>
  <div class="shelf">
    <UiEmptyState
      v-if="empty"
      :icon="icon"
      :title="lead"
      :text="hint"
    />

    <div
      v-else
      class="shelf__grid"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
// A shelf of cards, shared by sessions and instruments so both read alike.
defineProps<{
  empty?: boolean
  icon: string
  lead: string
  hint: string
}>()
</script>

<style scoped>
.shelf {
  min-height: 0;
  overflow-y: auto;
  padding: var(--gap);
}

.shelf__grid {
  display: grid;
  gap: var(--gap) var(--gap);
  grid-template-columns: repeat(auto-fill, minmax(135px, 1fr));
}
</style>
