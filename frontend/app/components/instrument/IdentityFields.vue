<template>
  <UiFieldGroup
    dense
    :title="t('instrument.sections.identity')"
  >
    <UiFormField
      v-slot="{ id }"
      :label="t('instrument.name')"
      :hint="t('instrument.nameHint')"
      :error="nameError ? t('instrument.nameRequired') : ''"
    >
      <UiTextInput
        :id="id"
        v-model="name"
        :invalid="nameError"
        :maxlength="MAX_INSTRUMENT_NAME"
        :placeholder="t('instrument.namePlaceholder')"
        @enter="emit('submit')"
      />
    </UiFormField>
  </UiFieldGroup>
</template>

<script setup lang="ts">
import { MAX_INSTRUMENT_NAME } from '~/types/session'

const name = defineModel<string>('name', { required: true })

defineProps<{ nameError: boolean }>()

const emit = defineEmits<{ submit: [] }>()

const { t } = useI18n()
</script>
