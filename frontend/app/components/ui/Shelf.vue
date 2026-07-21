<template>
  <div class="shelf">
    <!-- actions live in the toolbar, not here -->
    <div
      v-if="empty"
      class="shelf__empty"
    >
      <v-icon
        class="shelf__icon"
        :icon="icon"
      />
      <p class="shelf__lead">
        {{ lead }}
      </p>
      <p class="shelf__sub">
        {{ hint }}
      </p>
    </div>

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
  padding: 1cqh 0.8cqw;
}

.shelf__grid {
  display: grid;
  gap: 1cqh 0.8cqw;
  grid-template-columns: repeat(auto-fill, minmax(14cqw, 1fr));
}

.shelf__empty {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 1cqh;
  justify-content: center;
  min-height: 60cqh;
  text-align: center;
}

.shelf__icon {
  color: rgb(var(--v-theme-ink3));
  font-size: 7cqh;
  opacity: 0.4;
}

.shelf__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 2.1cqh;
  margin: 0;
}

.shelf__sub {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  line-height: 1.5;
  margin: 0;
  max-width: 48ch;
}
</style>
