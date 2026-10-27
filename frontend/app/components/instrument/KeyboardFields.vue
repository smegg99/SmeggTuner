<template>
  <UiFormSection
    :title="t('instrument.sections.keyboard')"
    class="tuner-section"
  >
    <UiFormField
      :label="t('instrument.range.label')"
      :hint="t('instrument.range.hint')"
    >
      <div class="form__range">
        <InstrumentNoteStepper v-model="lo" />
        <span class="form__rangeto">{{ t('instrument.range.to') }}</span>
        <InstrumentNoteStepper v-model="hi" />
      </div>
    </UiFormField>

    <p
      v-if="rangeError"
      class="form__warn"
    >
      {{ t('instrument.range.backwards') }}
    </p>
  </UiFormSection>
</template>

<script setup lang="ts">
// MIDI note numbers; 0 means the end is unset.
const lo = defineModel<number>('lo', { required: true })
const hi = defineModel<number>('hi', { required: true })

defineProps<{ rangeError: boolean }>()

const { t } = useI18n()
</script>

<style scoped>
.form__range {
  align-items: center;
  display: flex;
  gap: var(--gap);
}

.form__rangeto {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
}

.form__warn {
  color: rgb(var(--v-theme-warn));
  font-size: 11px;
  line-height: 1.4;
  margin: 0;
}
@media (max-width: 559px) {
  .form__range { flex-wrap: wrap; }
}
</style>
