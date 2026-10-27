<!-- app/components/session/SessionCardFace.vue -->
<template>
  <!-- The card's face: the photograph, the session and its accordion; stretched over the card, it opens it. -->
  <button
    type="button"
    class="card__open"
    :title="t('workshop.open')"
    @click="emit('open')"
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
      <span class="card__name">{{ name }}</span>
      <span class="card__what">{{ instrument || t('session.card.unnamedInstrument') }}</span>
    </span>
  </button>
</template>

<script setup lang="ts">
defineProps<{
  name: string
  instrument: string
  /** photograph url, or "" when none */
  src: string
}>()
const emit = defineEmits<{ open: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.card__open {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  cursor: pointer;
  display: flex;
  font: inherit;
  gap: var(--gap);
  min-width: 0;
  padding: var(--gap);
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

.card__pic {
  align-items: center;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 38px;
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
  font-size: 17px;
}

.card__text {
  display: flex;
  flex-direction: column;
  gap: var(--gap-tight);
  min-width: 0;
}

.card__name {
  font-size: 11.5px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__what {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
