<template>
  <div class="head">
    <span class="head__title">{{ title }}</span>
    <slot name="chip" />

    <span class="head__sp" />

    <slot />

    <!-- Opens a help dialog, not a tooltip. -->
    <button
      v-if="helpTitle"
      type="button"
      class="head__info"
      :title="helpTitle"
      :aria-label="helpTitle"
      @click="emit('help')"
    >
      <v-icon
        icon="mdi-information-outline"
        size="small"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
// Shared header strip so panels wear the same head.
defineProps<{ title: string, helpTitle?: string }>()
const emit = defineEmits<{ help: [] }>()
</script>

<style scoped>
.head {
  align-items: center;
  background: rgb(var(--v-theme-chrome));
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 7.7px;
  /* Fixed height so every panel's header lines up. */
  min-height: 26.8px;
  padding: 5.6px 9.9px;
}

.head__title {
  color: rgb(var(--v-theme-ink3));
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.head__sp {
  flex: 1 1 auto;
}

.head__info {
  align-items: center;
  background: none;
  border: 0;
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: inline-flex;
  padding: 1.6px 3.3px;
}

.head__info:hover {
  color: rgb(var(--v-theme-ink));
}

.head__info:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink2));
  outline-offset: 1px;
}
</style>
