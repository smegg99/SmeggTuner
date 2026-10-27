<template>
  <!-- A4 and judging windows are per-instrument, never global. -->
  <UiFormSection
    :title="t('instrument.sections.reference')"
    class="tuner-section"
  >
    <UiFormField
      v-show="mode !== 'tolerance'"
      v-slot="{ id }"
      :label="t('instrument.a4')"
      :hint="a4Note || t('instrument.a4Hint')"
      :error="a4Note"
    >
      <v-text-field
        :id="id"
        v-model="a4Draft"
        :suffix="t('settings.a4Unit')"
        :disabled="a4Disabled"
        hide-details="auto"
        autocomplete="off"
        :aria-label="t('instrument.a4')"
        inputmode="decimal"
        class="tuner-number"
        @blur="emit('commitA4')"
        @keyup.enter="emit('commitA4')"
      />
    </UiFormField>

    <div
      v-show="mode !== 'pitch'"
      class="form__row"
    >
      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.tolerance')"
        :hint="t('instrument.toleranceHint')"
      >
        <v-text-field
          :id="id"
          v-model="tolDraft"
          :suffix="t('settings.toleranceUnit')"
          hide-details="auto"
          autocomplete="off"
          :aria-label="t('instrument.tolerance')"
          inputmode="decimal"
          class="tuner-number"
          @blur="emit('commitTol')"
          @keyup.enter="emit('commitTol')"
        />
      </UiFormField>

      <UiFormField
        v-slot="{ id }"
        :label="t('instrument.beatTolerance')"
      >
        <v-text-field
          :id="id"
          v-model="beatDraft"
          :suffix="t('settings.toleranceUnit')"
          hide-details="auto"
          autocomplete="off"
          :aria-label="t('instrument.beatTolerance')"
          inputmode="decimal"
          class="tuner-number"
          @blur="emit('commitBeat')"
          @keyup.enter="emit('commitBeat')"
        />
      </UiFormField>
    </div>
  </UiFormSection>
</template>

<script setup lang="ts">
// Text drafts, clamped by the parent on commit (not mid-keystroke, so 442 can be typed through 44).
const a4Draft = defineModel<string>('a4Draft', { required: true })
const tolDraft = defineModel<string>('tolDraft', { required: true })
const beatDraft = defineModel<string>('beatDraft', { required: true })

// The reference pitch freezes once the session has readings; the field says so rather
// than offering a control the backend would refuse.
withDefaults(defineProps<{
  mode?: 'all' | 'pitch' | 'tolerance'
  a4Disabled?: boolean
  a4Note?: string
}>(), { mode: 'all' })

const emit = defineEmits<{ commitA4: [], commitTol: [], commitBeat: [] }>()

const { t } = useI18n()
</script>

<style scoped>
.form__row {
  display: grid;
  gap: var(--gap);
  grid-template-columns: 1fr 1fr;
}
</style>
