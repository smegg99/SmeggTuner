<template>
  <div class="list">
    <div
      v-if="items.length"
      class="list__grid"
    >
      <div
        v-for="{ item: it, index: i } in shown"
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

    <UiPageStepper
      v-if="compact && items.length > 1"
      v-model="page"
      :count="items.length"
      :label="t('instrument.sections.voices')"
    />
    <p
      v-if="!items.length"
      class="list__empty"
    >
      {{ empty }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
// One switch list for both keyboards: the treble section fills the symbol slot, the bass one doesn't.
const props = defineProps<{
  items: { name: string, feet: string }[]
  empty: string
}>()

const emit = defineEmits<{ remove: [index: number] }>()

const { t } = useI18n()
const compact = useCompactLayout()
const page = ref(0)
const shown = computed(() => props.items.map((item, index) => ({ item, index })).filter((_, index) => !compact.value || index === page.value))
watch(() => props.items.length, (length) => {
  page.value = Math.max(0, Math.min(page.value, length - 1))
})
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
  gap: var(--gap) var(--gap);
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
  gap: var(--gap);
  min-width: 0;
  padding: 4px 4px 4px 6.5px;
}

.item__sym {
  block-size: 28px;
  flex: 0 0 auto;
  inline-size: 28px;
}

/* The feet are the thing being read; they are never cut. */
.item__feet {
  color: rgb(var(--v-theme-ink));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 12px;
  white-space: nowrap;
}

/* The name is whatever was typed, so this is the one that gives way. */
.item__name {
  color: rgb(var(--v-theme-ink3));
  flex: 0 1 auto;
  max-width: 115.5px;
  font-family: var(--font-mono);
  font-size: 11px;
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
  padding-inline: 3px;
}

.item__kill:hover {
  color: rgb(var(--v-theme-error));
}

.list__empty {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  line-height: 1.5;
  margin: 0;
}
@media (max-width: 959px), (max-height: 559px) {
  .list { display: flex; gap: 6px; min-width: 0; align-items: center; }
  .list__grid { min-width: 0; flex: 1; }
  .item { max-width: 100%; flex: 1; padding-block: 0; }
  .item__name { flex: 1; }
  .item__kill { min-height: 40px; min-width: 32px; justify-content: center; }
  .list :deep(.pages) { flex: 0 0 auto; gap: 2px; }
  .list :deep(.pages__label) { display: none; }
}
</style>
