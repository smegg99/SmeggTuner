<template>
  <div class="list">
    <div
      v-if="items.length"
      class="list__grid"
    >
      <div
        v-for="(it, i) in items"
        :key="`${it.name}-${i}`"
        class="item"
      >
        <span
          v-if="$slots.symbol"
          class="item__sym"
        ><slot
          name="symbol"
          :index="i"
        /></span>
        <span class="item__feet">{{ it.feet }}</span>
        <span class="item__name">{{ it.name }}</span>
        <button
          type="button"
          class="item__kill"
          :title="t('instrument.builder.remove')"
          @click="emit('remove', i)"
        >
          <v-icon
            icon="mdi-close"
            size="x-small"
          />
        </button>
      </div>
    </div>

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
defineProps<{
  items: { name: string, feet: string }[]
  empty: string
}>()

const emit = defineEmits<{ remove: [index: number] }>()

const { t } = useI18n()
</script>

<style scoped>
/* A wrapping row, not a scrolling box. Switches are a handful of short labels: laid out like the
   presets they sit under, they all fit on screen at once and there is no inner scrollbar to hunt
   through. The panel scrolls if the whole form ever outgrows the pane.

   Sized to content rather than equal columns: a fixed column ellipsises "16'+8'+8'+8'" away, and
   that is the thing being read. */
.list__grid {
  display: flex;
  flex-wrap: wrap;
  gap: 4.8px 8.8px;
}

/* Metrics match InstrumentRegisterBuilder .preset: an added switch and an offered one are the
   same kind of thing and read as one set. */
.item {
  align-items: center;
  flex: 0 0 auto;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  gap: 5.5px;
  min-width: 0;
  padding: 4.4px 4.4px 4.4px 7.7px;
}

.item__sym {
  block-size: 32px;
  flex: 0 0 auto;
  inline-size: 32px;
}

/* The feet are the thing being read; they are never cut. */
.item__feet {
  color: rgb(var(--v-theme-ink));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 13.6px;
  white-space: nowrap;
}

/* The name is whatever was typed, so this is the one that gives way. */
.item__name {
  color: rgb(var(--v-theme-ink3));
  flex: 0 1 auto;
  max-width: 132px;
  font-family: var(--font-mono);
  font-size: 11.2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item__kill {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: flex;
  flex: 0 0 auto;
  padding-inline: 3.3px;
}

.item__kill:hover {
  color: rgb(var(--v-theme-error));
}

.list__empty {
  color: rgb(var(--v-theme-ink3));
  font-size: 12px;
  line-height: 1.5;
  margin: 0;
}
</style>
