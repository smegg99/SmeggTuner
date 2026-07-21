<template>
  <div class="card">
    <!-- The photo is a plain <img> against the asset server, never base64 in the bound object. -->
    <div class="card__pic">
      <img
        v-if="src"
        class="card__img"
        :src="src"
        :alt="instrument.name"
        loading="lazy"
      >
      <v-icon
        v-else
        class="card__none"
        icon="mdi-piano"
      />
    </div>

    <div class="card__body">
      <span class="card__name">{{ title }}</span>
      <span
        v-if="instrument.name && made"
        class="card__what"
      >{{ made }}</span>

      <!-- Banks in card order: which rank each reading files under. -->
      <div class="card__banks">
        <span
          v-for="b in banks"
          :key="b"
          class="card__bank"
        >{{ b }}</span>
        <span
          v-if="!banks.length"
          class="card__bank card__bank--none"
        >{{ t('instrument.card.noBanks') }}</span>
      </div>

      <span class="card__meta">
        {{ t('instrument.card.registers', registers) }}
      </span>
    </div>

    <div class="card__keys">
      <UiToolKey
        icon="mdi-camera-outline"
        :title="t('instrument.photo')"
        @click="emit('photo', instrument.id)"
      />
      <UiToolKey
        icon="mdi-pencil-outline"
        :title="t('instrument.edit')"
        @click="emit('edit', instrument.id)"
      />
      <UiToolKey
        icon="mdi-export"
        :title="t('instrument.export')"
        @click="emit('export', instrument.id)"
      />
      <UiToolKey
        icon="mdi-delete-outline"
        :title="t('instrument.delete')"
        @click="emit('delete', instrument.id)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { InstrumentTemplate } from '~/types/session'

const props = defineProps<{
  instrument: InstrumentTemplate
  // Photo URL, or "" when none.
  src: string
}>()

const emit = defineEmits<{
  edit: [id: string]
  photo: [id: string]
  export: [id: string]
  delete: [id: string]
}>()

const { t } = useI18n()

const made = computed(() => {
  const i = props.instrument.instrument
  return [i.make, i.model].filter(Boolean).join(' ')
})

// Heading falls back name -> make/model -> "not named"; the make/model line shows only when there's a distinct name above it.
const title = computed(() => props.instrument.name || made.value || t('instrument.card.unnamed'))

const banks = computed(() => props.instrument.instrument.banks ?? [])
const registers = computed(() => props.instrument.instrument.registers?.length ?? 0)
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
}

.card:hover {
  border-color: rgb(var(--v-theme-line));
}

.card__pic {
  align-items: center;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
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
  font-size: 5cqh;
  opacity: 0.5;
}

.card__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 0.3cqh;
  min-width: 0;
  padding: 1cqh 0.8cqw;
}

.card__name {
  color: rgb(var(--v-theme-ink));
  font-size: 1.7cqh;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__what {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.45cqh;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__banks {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25cqw;
  padding-block: 0.3cqh;
}

.card__bank {
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink2));
  font-family: var(--font-mono);
  font-size: 1.3cqh;
  padding: 0.15cqh 0.35cqw;
}

.card__bank--none {
  background: none;
  border-color: transparent;
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-sans);
}

.card__meta {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.35cqh;
}

.card__keys {
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 0.25cqw;
  padding: 0.6cqh 0.6cqw;
}
</style>
