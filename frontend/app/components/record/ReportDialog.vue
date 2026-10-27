<!-- app/components/record/ReportDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('record.report.title')"
    width="var(--dlg-md)"
    :persistent="busy"
  >
    <UiDialogCard
      :title="t('record.report.title')"
      :busy="busy"
      @close="open = false"
    >
      <p
        v-if="passLabel"
        class="mb-3"
      >
        {{ passLabel ? t('record.report.pass', { label: passLabel }) : '' }}
      </p>
      <div class="form">
        <div class="form__row">
          <UiFormField
            :label="t('record.report.format')"
            :hint="t('record.report.formatHint')"
          >
            <UiToolGroup
              v-model="draft.format"
              :items="FORMATS"
              :label="t('record.report.format')"
            />
          </UiFormField>

          <UiFormField
            :label="t('record.report.date')"
          >
            <template #default="{ id }">
              <v-text-field
                :id="id"
                v-model="draft.date"
                hide-details="auto"
                autocomplete="off"
                :aria-label="t('record.report.date')"
              />
            </template>
          </UiFormField>
        </div>

        <!-- The note is the one thing on the card that was not measured. It is written
             here because here is where the card gets printed. -->
        <UiFormField
          v-slot="{ id }"
          :label="t('record.report.notes')"
          :hint="t('record.report.notesHint')"
        >
          <v-textarea
            :id="id"
            v-model="notes"
            :rows="3"
            :disabled="busy"
            hide-details="auto"
            autocomplete="off"
            :aria-label="t('record.report.notes')"
          />
        </UiFormField>

        <!-- Letterhead lives in Settings; only per-export fields here. -->
        <p
          v-if="formError"
          class="form__error"
        >
          {{ formError }}
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
          :disabled="busy || notesTooLong"
          @click="submit"
        >
          {{ t('record.report.export') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { MAX_NOTES } from '~/utils/accordion'
import { useSessions } from '~/composables/useSessions'
import type { ToolItem } from '@smegg99/s99ab/types'
import type { ReportOptions } from '~/types/record'

// The backend writes the report from the session/pass it holds; renders none of the pass here.

const props = defineProps<{
  /** The pass being reported on, for the subtitle. */
  passLabel?: string
  busy?: boolean
  /** A failed export, already translated. */
  error?: string
}>()

const emit = defineEmits<{ export: [options: ReportOptions] }>()

const open = defineModel<boolean>({ default: false })

const { t } = useI18n()
// The note rides on the session, not on the export: Go reads it off the session it holds.
const { active, setNotes, error: sessionError } = useSessions()

const notes = ref('')
const savingNotes = ref(false)
const busy = computed(() => savingNotes.value || props.busy)
const saveError = ref('')
// Match Go's rune limit; native maxlength counts UTF-16 code units instead.
const notesTooLong = computed(() => Array.from(notes.value).length > MAX_NOTES)
const formError = computed(() => notesTooLong.value
  ? t('session.error.notesTooLong')
  : saveError.value ? t(saveError.value) : props.error)

watch(notes, () => {
  saveError.value = ''
})

// PDF first and default: same sheet as the HTML, laid out A4 by a headless browser on the backend.
const FORMATS = computed<ToolItem[]>(() => [
  { value: 'pdf', label: t('record.report.pdf'), icon: 'mdi-file-pdf-box' },
  { value: 'html', label: t('record.report.html'), icon: 'mdi-file-document-outline' },
  { value: 'csv', label: t('record.report.csv'), icon: 'mdi-file-delimited-outline' },
])

function today() {
  const now = new Date()
  const month = `${now.getMonth() + 1}`.padStart(2, '0')
  const day = `${now.getDate()}`.padStart(2, '0')
  return `${now.getFullYear()}-${month}-${day}`
}

const draft = reactive<ReportOptions>({ format: 'pdf', date: today() })

// Seed on open, not on mount: a reopened dialog would otherwise send a stale date and
// show a note that has since been rewritten.
watch(open, (showing) => {
  if (!showing) return
  saveError.value = ''
  draft.date = today()
  notes.value = active.value?.notes ?? ''
}, { immediate: true })

async function submit() {
  if (busy.value || notesTooLong.value) return
  savingNotes.value = true
  saveError.value = ''
  try {
    if (!await setNotes(notes.value)) {
      saveError.value = sessionError.value || 'session.error.unexpected'
      return
    }
    emit('export', { ...draft })
  }
  finally {
    savingNotes.value = false
  }
}
</script>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 154px;
  align-items: start;
  gap: 16px;
}

.form__error {
  color: rgb(var(--v-theme-reed));
  font-size: 11px;
  margin: 0;
}
</style>
