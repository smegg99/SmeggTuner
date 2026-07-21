<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-md)"
  >
    <UiDialogCard
      :title="t('record.report.title')"
      :subtitle="passLabel ? t('record.report.pass', { label: passLabel }) : ''"
      :confirm-label="t('record.report.export')"
      :loading="busy"
      @close="open = false"
      @cancel="open = false"
      @confirm="submit"
    >
      <div class="form">
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

        <UiFormField :label="t('record.report.date')">
          <template #default="{ id }">
            <UiTextInput
              :id="id"
              v-model="draft.date"
              class="form__date"
            />
          </template>
        </UiFormField>

        <!-- Letterhead lives in Settings; only per-export fields here. -->
        <p
          v-if="error"
          class="form__error"
        >
          {{ error }}
        </p>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { ReportOptions } from '~/types/record'

// The backend writes the report from the session/pass it holds; renders none of the pass here.

defineProps<{
  /** The pass being reported on, for the subtitle. */
  passLabel?: string
  busy?: boolean
  /** A failed export, already translated. */
  error?: string
}>()

const emit = defineEmits<{ export: [options: ReportOptions] }>()

const open = defineModel<boolean>({ default: false })

const { t } = useI18n()

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

// Seed the date on open, not on mount: a reopened dialog would otherwise send a stale date.
watch(open, (showing) => {
  if (showing) draft.date = today()
}, { immediate: true })

function submit() {
  emit('export', { ...draft })
}
</script>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 1.6cqh;
}

/* Narrow on purpose: a date is only 8 characters. */
.form__date {
  max-width: 16cqw;
}

.form__error {
  color: rgb(var(--v-theme-reed));
  font-size: 1.6cqh;
  margin: 0;
}
</style>
