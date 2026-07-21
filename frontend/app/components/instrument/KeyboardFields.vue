<template>
  <UiFieldGroup
    dense
    :title="t('instrument.sections.keyboard')"
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
  </UiFieldGroup>
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
  gap: 1cqw;
}

.form__rangeto {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
}

.form__warn {
  color: rgb(var(--v-theme-warn));
  font-size: 1.5cqh;
  line-height: 1.4;
  margin: 0;
}
</style>
