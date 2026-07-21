<template>
  <!-- A4 and judging windows are per-instrument, never global. -->
  <UiFieldGroup
    dense
    :title="t('instrument.sections.reference')"
  >
    <UiFormField
      v-slot="{ id }"
      :label="t('instrument.a4')"
      :hint="t('instrument.a4Hint')"
    >
      <UiNumberInput
        :id="id"
        v-model="a4Draft"
        :suffix="t('settings.a4Unit')"
        @commit="emit('commitA4')"
      />
    </UiFormField>

    <div class="form__row">
      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.tolerance')"
        :hint="t('instrument.toleranceHint')"
      >
        <UiNumberInput
          :id="id"
          v-model="tolDraft"
          :suffix="t('settings.toleranceUnit')"
          @commit="emit('commitTol')"
        />
      </UiFormField>

      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.beatTolerance')"
      >
        <UiNumberInput
          :id="id"
          v-model="beatDraft"
          :suffix="t('settings.toleranceUnit')"
          @commit="emit('commitBeat')"
        />
      </UiFormField>
    </div>
  </UiFieldGroup>
</template>

<script setup lang="ts">
// Text drafts, clamped by the parent on commit (not mid-keystroke, so 442 can be typed through 44).
const a4Draft = defineModel<string>('a4Draft', { required: true })
const tolDraft = defineModel<string>('tolDraft', { required: true })
const beatDraft = defineModel<string>('beatDraft', { required: true })

const emit = defineEmits<{ commitA4: [], commitTol: [], commitBeat: [] }>()

const { t } = useI18n()
</script>

<style scoped>
.form__row {
  display: grid;
  gap: 1.2cqw;
  grid-template-columns: 1fr 1fr;
}
</style>
