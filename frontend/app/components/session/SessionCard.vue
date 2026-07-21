<template>
  <div class="card">
    <!-- Whole card opens; the export/delete keys ride above the stretched hit area. -->
    <button
      type="button"
      class="card__open"
      :title="t('workshop.open')"
      @click="emit('open', session.id)"
    >
      <span class="card__pic">
        <img
          v-if="src"
          class="card__img"
          :src="src"
          :alt="instrument"
          loading="lazy"
        >
        <v-icon
          v-else
          class="card__none"
          icon="mdi-folder-music-outline"
        />
      </span>

      <span class="card__text">
        <span class="card__name">{{ session.name }}</span>
        <span class="card__what">{{ instrument || t('session.card.unnamedInstrument') }}</span>
      </span>
    </button>

    <v-tooltip
      v-if="changed"
      :text="t('session.card.changed')"
      location="left"
    >
      <template #activator="{ props: tip }">
        <v-icon
          v-bind="tip"
          class="card__alert"
          icon="mdi-alert-circle-outline"
          size="small"
        />
      </template>
    </v-tooltip>

    <SessionCardMeta :session="session" />

    <div class="card__keys">
      <span class="card__count">{{ t('session.card.readings', session.readings) }}</span>
      <span class="card__sp" />
      <UiToolKey
        icon="mdi-export"
        :title="t('session.export')"
        @click="emit('export', session.id)"
      />
      <UiToolKey
        icon="mdi-delete-outline"
        :title="t('session.delete')"
        @click="emit('delete', session.id)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import SessionCardMeta from './SessionCardMeta.vue'
import { useInstruments } from '~/composables/useInstruments'
import type { Summary } from '~~bindings/smegg.me/smeggtuner/core/session/models.js'

const props = defineProps<{
  session: Summary
  /** photograph url, or "" when none */
  src: string
}>()

const emit = defineEmits<{
  open: [id: string]
  export: [id: string]
  delete: [id: string]
}>()

const { t } = useI18n()
const { find } = useInstruments()

const instrument = computed(() => {
  const i = props.session.instrument
  return i.name || [i.make, i.model].filter(Boolean).join(' ')
})

// True when the shelf instrument's tuning-relevant fields diverge from the recorded snapshot; a renamed or missing instrument isn't flagged.
const changed = computed(() => {
  const id = props.session.instrumentId
  const shelf = id ? find(id)?.instrument : undefined
  return !!shelf && tuningSpec(shelf) !== tuningSpec(props.session.instrument)
})

function tuningSpec(i: {
  a4?: number
  tolerance?: number
  beatTolerance?: number
  banks?: unknown
  registers?: unknown
}): string {
  return JSON.stringify([i.a4 ?? 0, i.tolerance ?? 0, i.beatTolerance ?? 0, i.banks ?? [], i.registers ?? []])
}
</script>

<style scoped>
.card {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  position: relative;
  transition: background-color 120ms ease, border-color 120ms ease;
}

.card:hover {
  background: rgb(var(--v-theme-row));
  border-color: rgb(var(--v-theme-line));
}

/* Focus lands on the open button; the ring is drawn round the whole card. */
.card:has(.card__open:focus-visible) {
  outline: 2px solid rgb(var(--v-theme-ink2));
  outline-offset: 1px;
}

.card__open {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  cursor: pointer;
  display: flex;
  font: inherit;
  gap: 0.6cqw;
  min-width: 0;
  padding: 0.8cqh 0.7cqw;
  text-align: start;
  width: 100%;
}

/* Stretched over the card so any non-key area opens it. */
.card__open::after {
  content: "";
  inset: 0;
  position: absolute;
}

.card__open:focus-visible {
  outline: none;
}

/* Above the open button (z-index 1), tucked in the corner so it never crowds the name. */
.card__alert {
  color: rgb(var(--v-theme-warn));
  opacity: 0.7;
  position: absolute;
  right: 0.5cqw;
  top: 0.5cqh;
  z-index: 2;
}

.card__pic {
  align-items: center;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 5.4cqh;
  justify-content: center;
  overflow: hidden;
}

.card__img {
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.card__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 2.4cqh;
}

.card__text {
  display: flex;
  flex-direction: column;
  gap: 0.15cqh;
  min-width: 0;
}

.card__name {
  font-size: 1.65cqh;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__what {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__sp {
  flex: 1 1 auto;
}

.card__keys {
  align-items: center;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 0.25cqw;
  padding: 0.4cqh 0.5cqw;
  /* Above the stretched hit area: these are their own actions, not "open". */
  position: relative;
  z-index: 1;
}

.card__count {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.35cqh;
  overflow: hidden;
  padding-inline-start: 0.2cqw;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
