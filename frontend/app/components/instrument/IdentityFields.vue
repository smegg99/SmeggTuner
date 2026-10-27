<template>
  <UiFormSection
    :title="t('instrument.sections.identity')"
    class="tuner-section"
  >
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
        :error="nameError"
        @keyup.enter="emit('submit')"
      />
    </UiFormField>
  </UiFormSection>
</template>

<script setup lang="ts">
import { MAX_INSTRUMENT_NAME } from '~/types/session'

const name = defineModel<string>('name', { required: true })

defineProps<{ nameError: boolean }>()

const emit = defineEmits<{ submit: [] }>()

const { t } = useI18n()
</script>
