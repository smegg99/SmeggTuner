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

    <div class="form__row">
      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.make')"
      >
        <UiTextInput
          :id="id"
          v-model="make"
        />
      </UiFormField>

      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.model')"
      >
        <UiTextInput
          :id="id"
          v-model="model"
        />
      </UiFormField>
    </div>
  </UiFieldGroup>
</template>

<script setup lang="ts">
const name = defineModel<string>('name', { required: true })
const make = defineModel<string>('make', { required: true })
const model = defineModel<string>('model', { required: true })

defineProps<{ nameError: boolean }>()

const emit = defineEmits<{ submit: [] }>()

const { t } = useI18n()
</script>

<style scoped>
.form__row {
  display: grid;
  gap: 1.2cqw;
  grid-template-columns: 1fr 1fr;
}
</style>
