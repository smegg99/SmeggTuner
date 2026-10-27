<!-- app/components/session/AccordionQuickForm.vue -->
<template>
  <div class="quick">
    <UiFormField
      v-slot="{ id }"
      :label="t('instrument.name')"
      :hint="t('instrument.nameHint')"
      :error="nameError ? t('instrument.nameRequired') : ''"
    >
      <v-text-field
        :id="id"
        v-model="name"
        :maxlength="MAX_INSTRUMENT_NAME"
        :placeholder="t('instrument.namePlaceholder')"
        hide-details="auto"
        autocomplete="off"
        :aria-label="t('instrument.name')"
        :aria-invalid="nameError || undefined"
        :error="nameError"
      />
    </UiFormField>

    <UiFormField
      v-slot="{ id }"
      :label="t('instrument.a4')"
      :hint="t('instrument.a4Hint')"
    >
      <v-text-field
        :id="id"
        v-model="a4Draft"
        :suffix="t('settings.a4Unit')"
        hide-details="auto"
        autocomplete="off"
        :aria-label="t('instrument.a4')"
        inputmode="decimal"
        class="tuner-number"
        @blur="emit('commitA4')"
        @keyup.enter="emit('commitA4')"
      />
    </UiFormField>
  </div>
</template>

<script setup lang="ts">
import { MAX_INSTRUMENT_NAME } from '~/types/session'

// Two things, and neither can be measured: what the technician calls this accordion, and the
// pitch it is tuned to. Its switches are described on the Accordion tab, where there is room to
// build them one at a time.
const emit = defineEmits<{ commitA4: [] }>()
defineProps<{ nameError?: boolean }>()

const name = defineModel<string>('name', { required: true })
const a4Draft = defineModel<string>('a4Draft', { required: true })

const { t } = useI18n()
</script>

<style scoped>
.quick {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 112px;
  align-items: start;
  gap: 16px;
}
</style>
