<template>
  <v-menu>
    <template #activator="{ props: menu }">
      <button
        :id="id"
        v-bind="menu"
        type="button"
        class="pick"
      >
        <span
          class="pick__value"
          data-test="current"
        >{{ currentLabel }}</span>
        <v-icon
          icon="mdi-menu-down"
          class="pick__caret"
        />
      </button>
    </template>

    <UiMenuSheet>
      <!-- Not UiMenuItem: a row carries its own forget key, and a button cannot hold a button. -->
      <div
        v-for="i in items"
        :key="i.id"
        class="row"
      >
        <button
          type="button"
          class="row__pick"
          :class="{ 'row__pick--on': i.id === model }"
          :data-test="`pick-${i.id}`"
          @click="model = i.id"
        >
          <span class="row__name">{{ i.name || t('session.card.unnamedInstrument') }}</span>
          <span class="row__feet">{{ feetOf(i.instrument.banks ?? []) }}</span>
        </button>

        <button
          type="button"
          class="row__forget"
          :data-test="`forget-${i.id}`"
          :title="t('session.dialog.forgetAccordion')"
          @click.stop="emit('remove', i.id)"
        >
          <v-icon
            icon="mdi-close"
            size="x-small"
          />
        </button>
      </div>

      <span
        v-if="items.length"
        class="rule"
      />

      <!-- Their own events, not model values: the model only ever holds a real accordion. -->
      <UiMenuItem
        :label="t('session.dialog.describeNew')"
        data-test="describe"
        @click="emit('describe')"
      />
      <UiMenuItem
        :label="t('session.dialog.loadAccordion')"
        data-test="load"
        @click="emit('load')"
      />
    </UiMenuSheet>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { feetOf } from '~/utils/feet'
import type { InstrumentTemplate } from '~/types/session'

const props = defineProps<{
  items: readonly InstrumentTemplate[]
  id?: string
}>()

const emit = defineEmits<{
  describe: []
  load: []
  remove: [id: string]
}>()

const model = defineModel<string>({ required: true })

const { t } = useI18n()

const currentLabel = computed(() => {
  const picked = props.items.find(i => i.id === model.value)
  if (!picked) return t('session.dialog.pickAccordion')
  return picked.name || t('session.card.unnamedInstrument')
})
</script>

<style scoped>
/* Metrics copied from UiSelectInput .select so the two fields line up in one form. */
.pick {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  display: flex;
  font-family: var(--font-sans);
  font-size: 14px;
  gap: 5.5px;
  height: 35.2px;
  min-width: 0;
  padding-inline: 7.7px;
  text-align: start;
  width: 100%;
}

.pick:hover {
  border-color: rgb(var(--v-theme-ink3));
}

.pick:focus-visible {
  border-color: rgb(var(--v-theme-ink2));
}

.pick__value {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pick__caret {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-size: 16px;
}

.row {
  align-items: center;
  display: flex;
  min-width: 0;
}

.row:hover {
  background: rgb(var(--v-theme-row));
}

/* Metrics copied from UiMenuItem .item: these rows sit in the same sheet. */
.row__pick {
  align-items: baseline;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink2));
  cursor: pointer;
  display: flex;
  flex: 1 1 auto;
  font-family: var(--font-sans);
  font-size: 13.6px;
  gap: 6.6px;
  min-width: 0;
  padding: 6.8px 9.9px;
  text-align: start;
}

.row:hover .row__pick,
.row__pick--on {
  color: rgb(var(--v-theme-ink));
}

.row__name {
  flex: 0 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row__feet {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 11.6px;
}

.row__forget {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: flex;
  flex: 0 0 auto;
  padding-inline: 7.7px;
}

.row__forget:hover {
  color: rgb(var(--v-theme-error));
}

.rule {
  background: rgb(var(--v-theme-lineSoft));
  block-size: 1px;
  margin-block: 2.8px;
}
</style>
