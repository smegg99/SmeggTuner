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

  <UiFormField
    :label="t('settings.report.logo')"
    :hint="t('settings.report.logoHint')"
  >
    <div class="logo">
      <span class="logo__path">{{ logoName || t('settings.report.noLogo') }}</span>

      <UiToolKey
        icon="mdi-image-outline"
        :label="t('settings.report.pickLogo')"
        @click="pick"
      />
      <UiToolKey
        icon="mdi-close"
        :disabled="!logo"
        :title="t('settings.report.clearLogo')"
        @click="logo = ''"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as ReportService from '~~bindings/smegg.me/smeggtuner/services/report/service.js'
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

const logo = computed({
  get: () => config.report?.logo_path ?? '',
  set: (v: string) => { config.report.logo_path = v },
})

const logoName = computed(() => logo.value.split(/[/\\]/).pop() ?? '')

// Holds only the path; services/report.Export reads the file at export and core/report.LoadLogo
// owns the rule for what may be embedded.
async function pick() {
  const path = await ReportService.PickLogo()
  if (path) logo.value = path
}
</script>

<style scoped>
.logo {
  align-items: center;
  display: flex;
  gap: 0.5cqw;
  min-width: 0;
}

.logo__path {
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink2));
  flex: 1 1 auto;
  font-size: 1.6cqh;
  min-width: 0;
  overflow: hidden;
  padding: 0.9cqh 0.7cqw;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
