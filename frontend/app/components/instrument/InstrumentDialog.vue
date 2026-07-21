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
            v-model:make="form.make"
            v-model:model="form.model"
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
          <KeyboardFields
            v-model:lo="form.lo"
            v-model:hi="form.hi"
            :range-error="rangeError"
          />
        </section>

        <section class="form__col">
          <UiFieldGroup
            dense
            :title="t('instrument.sections.voices')"
          >
            <!-- Hint and summary share one slot: once registers exist the hint stays as an invisible
                 spacer under the summary, so the block keeps its height and the pane never moves. -->
            <div class="form__lead">
              <p
                class="form__note"
                :class="{ 'form__note--ghost': banks.length }"
              >
                {{ t('instrument.registersHint') }}
              </p>
              <p
                v-if="banks.length"
                class="form__says"
              >
                {{ t('instrument.says', { banks: banks.join(' '), reeds: reedCount }, reedCount) }}
              </p>
            </div>

            <InstrumentRegisterSection v-model="registers" />
          </UiFieldGroup>
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

.form__lead {
  position: relative;
}

.form__note {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.45cqh;
  line-height: 1.4;
  margin: 0;
}

/* Kept in flow as a spacer once the summary is shown, so the block's height does not change. */
.form__note--ghost {
  visibility: hidden;
}

/* Same font as the hint it stands over: the summary is always shorter text, so at equal size it can
   never wrap past the hint's height and spill onto the pane. Emphasis comes from the brighter ink. */
.form__says {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.45cqh;
  inset: 0;
  line-height: 1.4;
  margin: 0;
  position: absolute;
}
</style>
