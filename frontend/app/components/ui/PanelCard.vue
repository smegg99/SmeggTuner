<template>
  <div class="panel">
    <div class="panel__head">
      <slot name="head">
        <span
          v-if="title"
          class="panel__title"
        >{{ title }}</span>

        <span class="panel__sp" />

        <span class="panel__meta">
          <slot name="meta" />
        </span>
      </slot>
    </div>

    <div class="panel__body">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
// The head slot replaces the whole header row for panels that put controls there.
defineProps<{ title?: string }>()
</script>

<style scoped>
.panel {
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 0.45cqh;
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
}

/* min-height, not fixed height: grows for a header carrying controls */
.panel__head {
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  flex: 0 0 auto;
  flex-wrap: wrap; /* controls must wrap, not clip off the right edge */
  gap: 0.5cqw;
  min-height: 3.35cqh;
  min-width: 0;
  padding: 0.3cqh 0.75cqw;
}

/* :slotted lends these to slot content (scoped CSS otherwise can't reach it) */
.panel__title,
:slotted(.panel__title) {
  color: rgb(var(--v-theme-ink2));
  flex: 0 1 auto;
  font-size: 1.6cqh;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel__sp,
:slotted(.panel__sp) {
  flex: 1 1 auto;
}

.panel__meta,
:slotted(.panel__meta) {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-size: 1.55cqh;
  white-space: nowrap;
}

.panel__meta :deep(b) {
  color: rgb(var(--v-theme-ink2));
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

/* flex column so the child gets a bounded height and its overflow-y can bite */
.panel__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
  position: relative;
}

.panel__body > * {
  flex: 1 1 auto;
  min-height: 0;
}
</style>
