<template>
  <UiToolGroup
    class="views"
    :class="{ 'views--marks': !labelled }"
    :model-value="view"
    :items="VIEWS"
    :label="t('common.routes.navigation')"
    @update:model-value="v => setView(v as View)"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useShell } from '~/composables/useShell'
import { WORKSHOP } from '~/utils/features'
import type { View } from '~/composables/useShell'

// The three rooms - not routes: engine, shelf and config live for the whole process.
const { labelled = true, readingFirst = false } = defineProps<{
  /** names beside the marks; the toolbar drops them when they do not fit */
  labelled?: boolean
  /** Help and About before the rooms: a landscape bar keeps the tuner and the settings at its far end */
  readingFirst?: boolean
}>()

const { t } = useI18n()
const { view, setView } = useShell()
const compact = useCompactLayout()

// The name stays as title/aria-label when only the mark shows.
const ROOMS = computed<ToolItem[]>(() => [
  { value: 'tune', label: t('common.routes.tune'), icon: 'mdi-tune-vertical', title: t('common.routes.tune') },
  { value: 'workshop', label: t('common.routes.workshop'), icon: 'mdi-hammer-wrench', title: t('common.routes.workshop') },
  { value: 'settings', label: t('common.routes.settings'), icon: 'mdi-cog-outline', title: t('common.routes.settings') },
].filter(v => WORKSHOP || v.value !== 'workshop'))

// A compact window has no room for a dialog over it: Help and About are rooms there, left as any other
// is, and always marks. A wide window opens them from its status bar (ShellStatusActions).
const READING = computed<ToolItem[]>(() => [
  { value: 'help', icon: 'mdi-help-circle-outline', title: t('common.about.helpTitle') },
  { value: 'about', icon: 'mdi-information-outline', title: t('common.about.aboutTitle') },
])

const VIEWS = computed<ToolItem[]>(() => {
  if (!compact.value) return ROOMS.value
  return readingFirst ? [...READING.value, ...ROOMS.value] : [...ROOMS.value, ...READING.value]
})
</script>

<style scoped>
.views--marks :deep(.key) {
  width: auto;
  padding-inline: 8px;
}

.views--marks :deep(.key__label) {
  display: none;
}

/* Squares in a compact bar, as the keys beside them. */
@media (max-width: 959px), (max-height: 559px) {
  .views--marks :deep(.key) {
    width: var(--s99ab-control-h);
    padding-inline: 0;
  }
}
</style>
