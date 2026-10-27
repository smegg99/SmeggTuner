<!-- app/components/session/SessionDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('session.dialog.newTitle')"
    width="var(--dlg-md)"
    :persistent="busy"
  >
    <UiDialogCard
      :title="t('session.dialog.newTitle')"
      :busy="busy"
      @close="open = false"
    >
      <div class="form">
        <UiFormField
          v-if="mode === 'pick'"

          :label="t('session.dialog.instrument')"
        >
          <SessionAccordionPicker
            v-model="templateId"
            :items="library.list.value"
            @describe="mode = 'new'"
            @load="loadFromFile"
            @remove="forget"
          />
        </UiFormField>

        <div
          v-else

          class="form__instrument"
        >
          <SessionAccordionQuickForm
            v-model:name="acc.name"
            v-model:a4-draft="a4Draft"
            :name-error="nameError"
            @commit-a4="commitA4"
          />

          <button
            type="button"
            class="form__back"
            @click="mode = 'pick'"
          >
            {{ t('session.dialog.backToList') }}
          </button>
        </div>

        <UiFormField
          v-slot="{ id }"
          :label="t('session.dialog.name')"
          :hint="t('session.dialog.nameHint')"
        >
          <!-- Enter submits (not blur), so focusing the picker above doesn't start the session. -->
          <v-text-field
            :id="id"
            :model-value="form.name"
            :placeholder="offeredName"
            hide-details="auto"
            autocomplete="off"
            :aria-label="t('session.dialog.name')"
            @update:model-value="typeName"
            @keyup.enter="submit()"
          />
        </UiFormField>
        <p
          v-if="error"
          class="form__error"
          role="alert"
        >
          {{ t(error) }}
        </p>
      </div>
      <template #actions>
        <v-btn
          :disabled="busy"
          @click="open = false"
        >
          {{ t('common.cancel') }}
        </v-btn><v-btn
          color="accent"
          :loading="busy"
          :disabled="busy"
          @click="submit"
        >
          {{ t('session.dialog.start') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import type { SessionDraft } from '~/types/session'
import { useSessionDraft } from '~/components/session/useSessionDraft'

const props = defineProps<{
  saving?: boolean
}>()

const emit = defineEmits<{
  submit: [draft: SessionDraft]
}>()

const open = defineModel<boolean>({ required: true })

const { t } = useI18n()
const {
  library, mode, templateId, acc, form, nameError, error, a4Draft, commitA4, offeredName, busy,
  typeName, loadFromFile, forget, submit,
} = useSessionDraft(open, () => props.saving ?? false, draft => emit('submit', draft))
</script>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form__back {
  align-self: flex-start;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  font-family: var(--font-sans);
  font-size: 11px;
  padding: 0;
}

.form__error {
  color: rgb(var(--v-theme-warn));
  font-size: 11px;
  margin: 0;
}

/* No underline: nothing else in the app has one, and the colour lift is how every other
   text control here says it is live. */
.form__back:hover {
  color: rgb(var(--v-theme-ink));
}

.form__instrument { display: flex; flex-direction: column; gap: 8px; }
</style>
