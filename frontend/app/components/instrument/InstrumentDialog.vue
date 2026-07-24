<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-lg)"
  >
    <UiDialogCard
      :title="t(editing ? 'instrument.editTitle' : 'instrument.newTitle')"
      :confirm-label="t('common.save')"
      :loading="saving"
      @close="open = false"
      @cancel="open = false"
      @confirm="submit"
    >
      <div class="form">
        <section class="form__col">
          <IdentityFields
            v-model:name="form.name"
            :name-error="nameError"
            @submit="submit"
          />
          <ReferenceFields
            v-model:a4-draft="a4Draft"
            v-model:tol-draft="tolDraft"
            v-model:beat-draft="beatDraft"
            @commit-a4="commitA4"
            @commit-tol="commitTol"
            @commit-beat="commitBeat"
          />
          <UiFieldGroup
            dense
            :title="t('instrument.sections.bass')"
            :hint="t('instrument.bass.hint')"
          >
            <InstrumentBassSection
              v-model:reeds="bassReeds"
              v-model:registers="bassRegisters"
            />
          </UiFieldGroup>
        </section>

        <section class="form__col">
          <UiFieldGroup
            dense
            :title="t('instrument.sections.voices')"
            :hint="t('instrument.registersHint')"
          >
            <!-- One reserved line: a subtitle until switches exist, then what they add up to. -->
            <p class="form__says">
              {{ banks.length ? t('instrument.says', { banks: banks.join(' '), reeds: reedCount }, reedCount) : t('instrument.registersSubtitle') }}
            </p>

            <InstrumentRegisterSection v-model="registers" />
          </UiFieldGroup>

          <KeyboardFields
            v-model:lo="form.lo"
            v-model:hi="form.hi"
            :range-error="rangeError"
          />
        </section>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import IdentityFields from './IdentityFields.vue'
import ReferenceFields from './ReferenceFields.vue'
import KeyboardFields from './KeyboardFields.vue'
import { useInstrumentForm } from './useInstrumentForm'
import type { InstrumentTemplate } from '~/types/session'

const props = defineProps<{
  instrument?: InstrumentTemplate | null
  saving?: boolean
}>()

const emit = defineEmits<{ submit: [i: InstrumentTemplate] }>()

const open = defineModel<boolean>({ required: true })

const { t } = useI18n()

const {
  form,
  editing,
  nameError,
  a4Draft,
  tolDraft,
  beatDraft,
  registers,
  bassReeds,
  bassRegisters,
  banks,
  reedCount,
  rangeError,
  commitA4,
  commitTol,
  commitBeat,
  submit,
} = useInstrumentForm(props, emit, open)
</script>

<style scoped>
.form {
  display: grid;
  gap: 1.8cqh 2.8cqw;
  grid-template-columns: 1fr 1.1fr;
}

.form__col {
  display: flex;
  flex-direction: column;
  gap: 1.2cqh;
  min-width: 0;
}

/* One line, reserved even when empty, so describing switches never reflows the pane below. */
.form__says {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.45cqh;
  height: 2.1cqh;
  line-height: 2.1cqh;
  margin: 0;
  overflow: hidden;
}
</style>
