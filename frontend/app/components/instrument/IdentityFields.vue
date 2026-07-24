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
        :placeholder="t('instrument.namePlaceholder')"
        @enter="emit('submit')"
      />
    </UiFormField>
  </UiFieldGroup>
</template>

<script setup lang="ts">
const name = defineModel<string>('name', { required: true })

defineProps<{ nameError: boolean }>()

const emit = defineEmits<{ submit: [] }>()

const { t } = useI18n()
</script>
