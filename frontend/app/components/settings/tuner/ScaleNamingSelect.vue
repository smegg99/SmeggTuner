<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.scaleNamingLabel')"
    :hint="t('settings.scaleNamingHint')"
  >
    <v-select
      :id="id"
      v-model="naming"
      :items="items"
      item-title="label"
      item-value="value"
      hide-details="auto"
      :aria-label="t('settings.scaleNamingLabel')"
      :list-props="{ class: 'tuner-menu' }"
    >
      <template #item="{ props: itemProps, item }">
        <v-list-item
          v-bind="itemProps"
          :append-icon="(item.value === naming) ? 'mdi-check' : undefined"
        />
      </template>
    </v-select>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'

// Namings the schema allows (common/config). In cdefgah, H is B natural and B is B flat.
type ScaleNaming = 'cdefgab' | 'cdefgah' | 'doremi' | 'polish'

const NAMINGS: readonly ScaleNaming[] = ['cdefgab', 'cdefgah', 'polish', 'doremi']

const { t } = useI18n()
const { config } = useConfigSync()

const items = computed(() => NAMINGS.map(value => ({
  value,
  label: t(`settings.scaleNamings.${value}`),
})))

// Display follows immediately (note panel and equalizer read this field); the engine reads it
// only at its next start.
const naming = computed<ScaleNaming>({
  get: () => {
    const current = config.tuner?.scale_naming
    return NAMINGS.includes(current as ScaleNaming) ? current as ScaleNaming : 'cdefgab'
  },
  set: (next: ScaleNaming) => {
    config.tuner.scale_naming = next
  },
})
</script>
