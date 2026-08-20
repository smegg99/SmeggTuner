<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.name')"
    :hint="t('settings.report.nameHint')"
  >
    <UiTextInput
      :id="id"
      v-model="name"
    />
  </UiFormField>

  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.address')"
  >
    <UiTextInput
      :id="id"
      v-model="address"
      :rows="2"
    />
  </UiFormField>

  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.website')"
  >
    <UiTextInput
      :id="id"
      v-model="website"
    />
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

const { t } = useI18n()
const { config } = useConfigSync()

const name = computed({
  get: () => config.report?.company_name ?? '',
  set: (v: string) => { config.report.company_name = v },
})

const address = computed({
  get: () => config.report?.company_address ?? '',
  set: (v: string) => { config.report.company_address = v },
})

const website = computed({
  get: () => config.report?.company_website ?? '',
  set: (v: string) => { config.report.company_website = v },
})
</script>
