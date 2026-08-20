<template>
  <div class="quick">
    <UiFormField
      v-slot="{ id }"
      :label="t('instrument.name')"
      :hint="t('instrument.nameHint')"
    >
      <UiTextInput
        :id="id"
        v-model="name"
        :maxlength="MAX_INSTRUMENT_NAME"
        :placeholder="t('instrument.namePlaceholder')"
      />
    </UiFormField>

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
  </div>
</template>

<script setup lang="ts">
import { MAX_INSTRUMENT_NAME } from '~/types/session'

// Two things, and neither can be measured: what the technician calls this accordion, and the
// pitch it is tuned to. Its switches are described on the Accordion tab, where there is room to
// build them one at a time.
const emit = defineEmits<{ commitA4: [] }>()

const name = defineModel<string>('name', { required: true })
const a4Draft = defineModel<string>('a4Draft', { required: true })

const { t } = useI18n()
</script>

<style scoped>
.quick {
  display: flex;
  flex-direction: column;
  gap: 11.2px;
}
</style>
