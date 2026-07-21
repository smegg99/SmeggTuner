<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.languageLabel')"
    :hint="t('settings.languageHint')"
  >
    <UiSelectInput
      :id="id"
      v-model="language"
      :items="languageItems"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SelectItem } from '~/components/ui/SelectInput.vue'
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

const languageItems = computed<SelectItem[]>(() =>
  localeItems.value.map(l => ({ value: l.code, label: l.name })),
)
</script>
