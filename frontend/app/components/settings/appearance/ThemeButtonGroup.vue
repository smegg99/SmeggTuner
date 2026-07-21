<template>
  <UiFormField
    :label="t('settings.themeLabel')"
    :hint="t('settings.themeHint')"
  >
    <UiToolGroup
      :model-value="mode"
      :items="themeItems"
      :label="t('settings.themeLabel')"
      @update:model-value="onSelect"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useThemeSync } from '~/composables/useThemeSync'
import { THEME_MODES } from '~/types/config'

const { t } = useI18n()
const { mode, setThemeMode } = useThemeSync()

// The toggle can hand back null while settling; only write a mode the config admits.
function onSelect(value: unknown) {
  const next = THEME_MODES.find(item => item.value === value)
  if (next) setThemeMode(next.value)
}

const themeItems = computed<ToolItem[]>(() =>
  THEME_MODES.map(m => ({ value: m.value, label: t(m.labelKey), icon: m.icon })),
)
</script>
