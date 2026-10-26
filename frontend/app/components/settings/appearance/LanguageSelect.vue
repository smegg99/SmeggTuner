<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.languageLabel')"
    :hint="t('settings.languageHint')"
  >
    <v-select
      :id="id"
      v-model="language"
      :items="languageItems"
      item-title="label"
      item-value="value"
      hide-details="auto"
      :aria-label="t('settings.languageLabel')"
      :list-props="{ class: 'tuner-menu' }"
    >
      <template #item="{ props: itemProps, item }">
        <v-list-item
          v-bind="itemProps"
          :append-icon="(item.value === language) ? 'mdi-check' : undefined"
        />
      </template>
    </v-select>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeSync } from '~/composables/useThemeSync'
import type { LocaleCode } from '~/types/locale'

const { prefs, localeItems, setLanguage } = useThemeSync()
const { t, locale } = useI18n()

const language = computed({
  get: () => prefs.value.language || locale.value || 'en',
  set: (v: string) => {
    void setLanguage(v as LocaleCode)
  },
})

const languageItems = computed(() =>
  localeItems.value.map(l => ({ value: l.code, label: l.name })),
)
</script>
