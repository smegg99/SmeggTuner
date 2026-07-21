<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-md)"
  >
    <UiDialogCard
      :title="t('session.dialog.newTitle')"
      :confirm-label="t('session.dialog.create')"
      :loading="saving"
      @close="open = false"
      @cancel="open = false"
      @confirm="submit"
    >
      <div class="form">
        <UiFormField
          v-slot="{ id }"
          :label="t('session.dialog.name')"
          :hint="t('session.dialog.nameHint')"
          :error="nameError ? t('session.dialog.nameRequired') : ''"
        >
          <!-- Enter submits (not blur), so focusing the picker below doesn't create the session. -->
          <UiTextInput
            :id="id"
            v-model="form.name"
            :invalid="nameError"
            :placeholder="t('session.dialog.namePlaceholder')"
            @enter="submit"
          />
        </UiFormField>

        <SessionInstrumentField
          v-model="templateId"
          :instrument="form.instrument"
          :items="instrumentItems"
          :disabled="library.loading.value || library.empty.value"
        />

        <UiFormField
          v-slot="{ id }"
          :label="t('session.dialog.notes')"
        >
          <UiTextInput
            :id="id"
            v-model="form.notes"
            :rows="3"
            :maxlength="500"
          />
        </UiFormField>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import SessionInstrumentField from './SessionInstrumentField.vue'
import { emptyInstrument } from '~/types/session'
import type { Instrument } from '~/types/session'

// Carries the whole instrument (ranks, switches, keyboard, reed count), not just the shown fields.
export interface SessionDraft {
  name: string
  notes: string
  instrument: Instrument
  /** shelf instrument id, so the session can find its photograph */
  instrumentId: string
}

defineProps<{
  saving?: boolean
}>()

const emit = defineEmits<{
  submit: [draft: SessionDraft]
}>()

const open = defineModel<boolean>({ required: true })

const { t } = useI18n()
const library = useInstruments()

const nameError = ref(false)
const templateId = ref('')
const form = reactive<SessionDraft>(blank())

const instrumentItems = computed(() => [
  { value: '', label: t('session.dialog.instrumentNone') },
  ...library.list.value.map(i => ({ value: i.id, label: i.name })),
])

function blank(): SessionDraft {
  return {
    name: '',
    notes: '',
    instrument: emptyInstrument(),
    instrumentId: '',
  }
}

// Reset on every open so a cancelled draft never leaks into the next session.
watch(open, async (isOpen) => {
  if (!isOpen) return

  nameError.value = false
  await library.load()

  Object.assign(form, blank())
  templateId.value = ''
})

watch(templateId, (id) => {
  if (!open.value) return

  if (!id) {
    form.instrument = emptyInstrument()
    return
  }

  const picked = library.find(id)
  if (!picked) return

  // Name lives on the template, not the core instrument; copy it or a named make/model-less box shows as unnamed.
  form.instrument = { ...picked.instrument, name: picked.name }
})

function submit() {
  const name = form.name.trim()
  nameError.value = name === ''
  if (nameError.value) return

  emit('submit', { ...form, name, instrumentId: templateId.value, instrument: { ...form.instrument } })
}
</script>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 2.4cqh;
}
</style>
