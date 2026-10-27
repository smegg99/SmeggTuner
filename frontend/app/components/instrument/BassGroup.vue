<!-- app/components/instrument/BassGroup.vue -->
<template>
  <!-- The bass machine's section of the accordion form, its hint on the heading. -->
  <UiFormSection
    :title="t('instrument.sections.bass')"
    class="tuner-section"
  >
    <template #head>
      <span class="form-section__title section-hint">
        {{ t('instrument.sections.bass') }}
        <v-tooltip
          activator="parent"
          location="top start"
          max-width="340"
          open-delay="300"
        >
          {{ t('instrument.bass.hint') }}
        </v-tooltip>
      </span>
    </template>
    <InstrumentBassSection
      v-model:reeds="reeds"
      v-model:registers="registers"
      :mode="mode"
    />
  </UiFormSection>
</template>

<script setup lang="ts">
import type { BassRegister } from '~/types/session'

const reeds = defineModel<number>('reeds', { required: true })
const registers = defineModel<BassRegister[]>('registers', { required: true })
defineProps<{ mode: 'all' | 'list' | 'builder' }>()
const { t } = useI18n()
</script>
