<template>
  <div
    class="list"
    :style="{ height }"
  >
    <ul
      v-if="items.length"
      class="list__rows"
    >
      <li
        v-for="(it, i) in items"
        :key="`${it.name}-${i}`"
        class="list__row"
      >
        <span
          v-if="$slots.symbol"
          class="list__sym"
        ><slot
          name="symbol"
          :index="i"
        /></span>
        <span class="list__feet">{{ it.feet }}</span>
        <span class="list__name">{{ it.name }}</span>
        <button
          type="button"
          class="list__kill"
          :title="t('instrument.builder.remove')"
          @click="emit('remove', i)"
        >
          <v-icon
            icon="mdi-close"
            size="x-small"
          />
        </button>
      </li>
    </ul>

    <p
      v-else
      class="list__empty"
    >
      {{ empty }}
    </p>
  </div>
</template>

<script setup lang="ts">
// One switch list for both keyboards: the treble section fills the symbol slot, the bass one doesn't.
withDefaults(defineProps<{
  items: { name: string, feet: string }[]
  empty: string
  height?: string
}>(), { height: '30cqh' })

const emit = defineEmits<{ remove: [index: number] }>()

const { t } = useI18n()
</script>

<style scoped>
/* Fixed height so adding registers scrolls inside the pane instead of growing the dialog; gutter reserved so the remove key never collides with the scrollbar. */
.list {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  overflow-y: auto;
  padding: 0.6cqh 0.5cqw;
  scrollbar-gutter: stable;
}

.list__rows {
  display: flex;
  flex-direction: column;
  gap: 0.4cqh;
  list-style: none;
  margin: 0;
  padding: 0;
}

.list__row {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  gap: 0.8cqw;
  padding: 0.7cqh 0.7cqw;
}

.list__sym {
  block-size: 5cqh;
  flex: 0 0 auto;
  inline-size: 5cqh;
}

.list__feet {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 1.9cqh;
}

.list__name {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-mono);
  font-size: 1.5cqh;
}

.list__kill {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: flex;
  margin-inline-start: auto;
}

.list__kill:hover {
  color: rgb(var(--v-theme-error));
}

.list__empty {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  line-height: 1.5;
  margin: 0;
}
</style>
