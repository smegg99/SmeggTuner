<template>
  <div class="card">
    <!-- Whole card opens; the export/delete keys ride above the stretched hit area. -->
    <SessionCardFace
      :name="session.name"
      :instrument="instrument"
      :src="src"
      @open="emit('open', session.id)"
    />

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
import SessionCardFace from './SessionCardFace.vue'
import SessionCardMeta from './SessionCardMeta.vue'
import { useInstruments } from '~/composables/useInstruments'
import type { Summary } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'

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
  return i.name ?? ''
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

/* Above the open button (z-index 1), tucked in the corner so it never crowds the name. */
.card__alert {
  color: rgb(var(--v-theme-warn));
  opacity: 0.7;
  position: absolute;
  right: 5px;
  top: 3.5px;
  z-index: 2;
}

.card__sp {
  flex: 1 1 auto;
}

.card__keys {
  align-items: center;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: var(--gap-tight);
  padding: 3px 5px;
  /* Above the stretched hit area: these are their own actions, not "open". */
  position: relative;
  z-index: 1;
}

.card__count {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  overflow: hidden;
  padding-inline-start: var(--gap-tight);
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
