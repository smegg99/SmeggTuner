<template>
  <UiToolGroup
    class="views"
    :model-value="view"
    :items="VIEWS"
    :label="t('routes.navigation')"
    @update:model-value="v => setView(v as View)"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useShell } from '~/composables/useShell'
import type { View } from '~/composables/useShell'

// The three rooms - not routes: engine, shelf and config live for the whole process.
const { t } = useI18n()
const { view, setView } = useShell()

// Names by default, dropped to marks when the row is tight (see CSS); the name stays as title/aria-label.
const VIEWS = computed<ToolItem[]>(() => [
  { value: 'tune', label: t('routes.tune'), icon: 'mdi-tune-vertical', title: t('routes.tune') },
  { value: 'workshop', label: t('routes.workshop'), icon: 'mdi-hammer-wrench', title: t('routes.workshop') },
  { value: 'settings', label: t('routes.settings'), icon: 'mdi-cog-outline', title: t('routes.settings') },
])
</script>

<style scoped>
/* Names collapse to marks when the row is tight. Measured threshold: 1180px clears Polish (the
   wider, decisive language) with room; re-measure if a longer language or a tune-row key is added.
   Hide the whole label, not just its text: UiToolKey ghosts every face to fix its width. */
@container (max-width: 1180px) {
  .views :deep(.key__label) {
    display: none;
  }
}
</style>
