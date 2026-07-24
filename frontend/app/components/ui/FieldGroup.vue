<template>
  <section
    class="group"
    :class="{ 'group--dense': dense }"
  >
    <h5
      class="group__title"
      :class="{ 'group__title--hinted': hint }"
    >
      {{ title }}<v-tooltip
        v-if="hint"
        activator="parent"
        location="top start"
        max-width="340"
        open-delay="300"
      >
        {{ hint }}
      </v-tooltip>
    </h5>

    <div class="group__fields">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  // Hovers on the title, so a section can explain itself without spending a paragraph of dialog.
  hint?: string
  dense?: boolean
}>()
</script>

<style scoped>
.group {
  display: flex;
  flex-direction: column;
  gap: 1.4cqh;
  min-width: 0;
}

.group--dense {
  gap: 1cqh;
}

.group__title {
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  font-weight: 700;
  letter-spacing: 0.06em;
  margin: 0;
  padding-bottom: 0.6cqh;
  text-transform: uppercase;
}

.group__title--hinted {
  cursor: help;
}

.group__fields {
  display: flex;
  flex-direction: column;
  gap: 1.6cqh;
  min-width: 0;
}

.group--dense .group__fields {
  gap: 0.9cqh;
}
</style>
