<!-- app/debug/PhoneLayout.vue -->
<!-- A debug build's setting: the window as a phone, upright or on its side, to test the layout the Android app shows. -->
<template>
  <UiFormField
    row
    :label="t('settings.phoneLayout.label')"
    :hint="t('settings.phoneLayout.hint')"
  >
    <UiToolGroup
      :model-value="phoneLayout"
      :items="items"
      :label="t('settings.phoneLayout.label')"
      @update:model-value="onSelect"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { PHONE_LAYOUTS, phoneLayout, setPhoneLayout } from './phone'

const { t } = useI18n()

// The toggle can hand back null while settling; only a known layout sizes the window.
function onSelect(value: unknown) {
  const next = PHONE_LAYOUTS.find(layout => layout.value === value)
  if (next) void setPhoneLayout(next.value)
}

const items = computed<ToolItem[]>(() =>
  PHONE_LAYOUTS.map(layout => ({ value: layout.value, label: t(layout.labelKey) })),
)
</script>
