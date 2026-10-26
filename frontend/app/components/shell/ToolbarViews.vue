<template>
  <UiToolGroup
    class="views"
    :class="{ 'views--marks': !labelled }"
    :model-value="view"
    :items="VIEWS"
    :label="t('routes.navigation')"
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
const { labelled = true } = defineProps<{
  /** names beside the marks; the toolbar drops them when they do not fit */
  labelled?: boolean
}>()

const { t } = useI18n()
const { view, setView } = useShell()

// The name stays as title/aria-label when only the mark shows.
const VIEWS = computed<ToolItem[]>(() => [
  { value: 'tune', label: t('routes.tune'), icon: 'mdi-tune-vertical', title: t('routes.tune') },
  { value: 'workshop', label: t('routes.workshop'), icon: 'mdi-hammer-wrench', title: t('routes.workshop') },
  { value: 'settings', label: t('routes.settings'), icon: 'mdi-cog-outline', title: t('routes.settings') },
].filter(v => WORKSHOP || v.value !== 'workshop'))
</script>

<style scoped>
.views--marks :deep(.key) {
  width: auto;
  padding-inline: 8px;
}

.views--marks :deep(.key__label) {
  display: none;
}
</style>
