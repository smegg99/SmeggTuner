<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.name')"
    :hint="t('settings.report.nameHint')"
  >
    <v-text-field
      :id="id"
      v-model="name"
      hide-details="auto"
      autocomplete="off"
      :aria-label="t('settings.report.name')"
    />
  </UiFormField>

  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.address')"
  >
    <v-textarea
      :id="id"
      v-model="address"
      :rows="2"
      hide-details="auto"
      autocomplete="off"
      :aria-label="t('settings.report.address')"
    />
  </UiFormField>

  <UiFormField
    v-slot="{ id }"
    :label="t('settings.report.website')"
  >
    <v-text-field
      :id="id"
      v-model="website"
      hide-details="auto"
      autocomplete="off"
      :aria-label="t('settings.report.website')"
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
