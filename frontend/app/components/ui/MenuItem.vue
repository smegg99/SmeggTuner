<template>
  <button
    type="button"
    class="item"
    :class="{ 'item--on': active }"
    :disabled="disabled"
    :title="label"
    :aria-current="active"
  >
    <v-icon
      v-if="icon"
      class="item__icon"
      :icon="icon"
    />

    <!-- external names (device, recording) can be long: ellipsise, full text on title -->
    <span class="item__label">{{ label }}</span>

    <v-icon
      v-if="active"
      class="item__tick"
      icon="mdi-check"
    />
  </button>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  icon?: string
  active?: boolean
  disabled?: boolean
}>()
</script>

<style scoped>
.item {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink2));
  cursor: pointer;
  display: flex;
  font-family: var(--font-sans);
  font-size: 13.6px;
  gap: 6.6px;
  min-width: 0;
  padding: 6.8px 9.9px;
  text-align: start;
  width: 100%;
}

.item:hover:not(:disabled) {
  background: rgb(var(--v-theme-row));
  color: rgb(var(--v-theme-ink));
}

.item:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: -2px;
}

.item:disabled {
  color: rgb(var(--v-theme-ink3));
  cursor: default;
}

.item--on {
  color: rgb(var(--v-theme-ink));
}

.item__label {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item__icon {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-size: 16px;
}

.item__tick {
  color: rgb(var(--v-theme-ink2));
  flex: 0 0 auto;
  font-size: 15.2px;
}
</style>
