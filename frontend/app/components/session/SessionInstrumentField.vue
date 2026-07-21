<template>
  <div class="ifield">
    <UiFormField
      v-slot="{ id }"
      :label="t('session.dialog.instrument')"
      :hint="t('session.dialog.instrumentHint')"
    >
      <UiSelectInput
        :id="id"
        v-model="modelValue"
        :items="items"
        :disabled="disabled"
      />
    </UiFormField>

    <p class="form__says">
      {{ says }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Instrument } from '~/types/session'

const props = defineProps<{
  instrument: Instrument
  items: { value: string, label: string }[]
  disabled?: boolean
}>()

const modelValue = defineModel<string>({ required: true })

const { t } = useI18n()

const described = computed(() => (props.instrument.banks?.length ?? 0) > 0)

const says = computed(() => {
  if (!described.value) return t('session.dialog.saysNothing')
  const registers = props.instrument.registers?.length ?? 0
  return t('session.dialog.says', {
    banks: (props.instrument.banks ?? []).join(' '),
    registers,
  }, registers)
})
</script>

<style scoped>
/* display:contents so the field and summary stay direct flex children of .form. */
.ifield {
  display: contents;
}

.form__says {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.5cqh;
  margin: 0;
  max-width: 46ch; /* variable-length text must not grow the dialog */
  overflow-wrap: anywhere;
}
</style>
