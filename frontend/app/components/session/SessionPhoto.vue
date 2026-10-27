<!-- app/components/session/SessionPhoto.vue -->
<template>
  <!-- The photograph is the picker, but only where the accordion is edited. Everywhere else it
       is a plain span: no button, no tooltip, no hover, so nothing invites a click that would
       do nothing. It also needs a shelf entry to hang the image on, so it waits for one. -->
  <component
    :is="live ? 'button' : 'span'"
    class="ident__pic"
    :class="{ 'ident__pic--live': live }"
    :type="live ? 'button' : undefined"
    :title="live ? t('instrument.photo') : undefined"
    @click="live && emit('pick')"
  >
    <img
      v-if="src"
      class="ident__img"
      :src="src"
      :alt="alt"
    >
    <v-icon
      v-else
      :icon="live ? 'mdi-camera-outline' : 'mdi-folder-music-outline'"
      class="ident__none"
    />
  </component>
</template>

<script setup lang="ts">
defineProps<{
  /** a picker: only on the accordion tab, with a shelf entry to hang the image on */
  live: boolean
  /** photograph url, or "" when none */
  src: string
  alt: string
}>()
const emit = defineEmits<{ pick: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.ident__pic {
  align-items: center;
  padding: 0;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 35px;
  justify-content: center;
  overflow: hidden;
}

.ident__pic--live {
  cursor: pointer;
}

.ident__pic--live:hover {
  border-color: rgb(var(--v-theme-ink2));
}

.ident__pic--live:hover .ident__none {
  color: rgb(var(--v-theme-ink));
}

.ident__img {
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.ident__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 17px;
}
</style>
