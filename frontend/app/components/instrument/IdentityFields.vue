<template>
  <UiFormSection
    :title="t('session.instrument.sections.identity')"
    class="tuner-section"
  >
    <UiFormField
      v-slot="{ id }"
      :label="t('session.instrument.name')"
      :hint="t('session.instrument.nameHint')"
      :error="nameError ? t('session.instrument.nameRequired') : ''"
    >
      <v-text-field
        :id="id"
        v-model="name"
        :maxlength="MAX_INSTRUMENT_NAME"
        :placeholder="t('session.instrument.namePlaceholder')"
        hide-details="auto"
        autocomplete="off"
        :aria-label="t('session.instrument.name')"
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
