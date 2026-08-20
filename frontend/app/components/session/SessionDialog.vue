<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-md)"
  >
    <UiDialogCard
      :title="t('session.dialog.newTitle')"
      :confirm-label="t('session.dialog.start')"
      :confirm-disabled="!canStart"
      :loading="saving"
      @close="open = false"
      @cancel="open = false"
      @confirm="submit"
    >
      <div class="form">
        <UiFormField
          v-if="mode === 'pick'"
          v-slot="{ id }"
          :label="t('session.dialog.instrument')"
        >
          <SessionAccordionPicker
            :id="id"
            v-model="templateId"
            :items="library.list.value"
            @describe="mode = 'new'"
            @load="loadFromFile"
            @remove="forget"
          />
        </UiFormField>

        <template v-else>
          <SessionAccordionQuickForm
            v-model:name="acc.name"
            v-model:a4-draft="a4Draft"
            @commit-a4="commitA4"
          />

          <button
            type="button"
            class="form__back"
            @click="mode = 'pick'"
          >
            {{ t('session.dialog.backToList') }}
          </button>
        </template>

        <UiFormField
          v-slot="{ id }"
          :label="t('session.dialog.name')"
          :hint="t('session.dialog.nameHint')"
        >
          <!-- Enter submits (not blur), so focusing the picker above doesn't start the session. -->
          <UiTextInput
            :id="id"
            :model-value="form.name"
            :placeholder="offeredName"
            @update:model-value="typeName"
            @enter="submit"
          />
        </UiFormField>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { sessionNameFor } from '~/utils/accordion'
import { A4_DEFAULT, A4_MAX, A4_MIN } from '~/types/session'
import type { Instrument } from '~/types/session'

// Carries the whole instrument (ranks, switches, reference pitch), not just the shown fields.
export interface SessionDraft {
  name: string
  notes: string
  instrument: Instrument
  /** remembered accordion id, so the session can find its photograph */
  instrumentId: string
}

defineProps<{
  saving?: boolean
}>()

const emit = defineEmits<{
  submit: [draft: SessionDraft]
}>()

const open = defineModel<boolean>({ required: true })

const { t, d } = useI18n()
const library = useInstruments()

// Picking one already tuned, or describing the one on the bench. Never both.
const mode = ref<'pick' | 'new'>('pick')
const templateId = ref('')
const acc = reactive({ name: '', a4: A4_DEFAULT })
const form = reactive({ name: '' })
const typedName = ref(false)

const { draft: a4Draft, commit: commitA4 } = useDraftNumber(
  () => acc.a4,
  (hz) => { acc.a4 = hz },
  { min: A4_MIN, max: A4_MAX },
)

// The name follows whichever accordion is in hand, until it is typed over.
const offeredName = computed(() => sessionNameFor(
  mode.value === 'new' ? acc.name : (library.find(templateId.value)?.name ?? ''),
  d(new Date(), 'short'),
))

// Nothing gates the start any more: the session name is offered, the accordion's own name may be
// left blank, and its switches are described on the Accordion tab once the session is open.
const canStart = computed(() => true)

watch(offeredName, (next) => {
  if (!typedName.value) form.name = next
}, { immediate: true })

// Reset on every open so a cancelled draft never leaks into the next session.
watch(open, async (isOpen) => {
  if (!isOpen) return

  mode.value = 'pick'
  templateId.value = ''
  typedName.value = false
  acc.name = ''
  acc.a4 = A4_DEFAULT
  a4Draft.value = String(A4_DEFAULT)

  await library.refresh()
})

function typeName(next: string) {
  typedName.value = true
  form.name = next
}

async function loadFromFile() {
  const added = await library.importFile()
  if (added) templateId.value = added.id
}

async function forget(id: string) {
  await library.remove(id)
  if (templateId.value === id) templateId.value = ''
}

async function submit() {
  if (!canStart.value) return
  commitA4()

  const picked = mode.value === 'new' ? await remember() : library.find(templateId.value)
  if (mode.value === 'new' && !picked) return

  emit('submit', {
    name: form.name.trim() || offeredName.value,
    notes: '',
    instrument: picked
      ? { ...picked.instrument, name: picked.name }
      : { serial: '', reedCount: 1 },
    instrumentId: picked?.id ?? '',
  })
}

// An accordion is remembered as a side effect of starting work on it; there is no other way one
// gets into the list. It starts with no switches described - one reed, no ranks - and the
// Accordion tab is where they get added.
function remember() {
  const name = acc.name.trim()

  return library.save({
    id: '',
    name,
    hasImage: false,
    instrument: { name, serial: '', reedCount: 1, a4: acc.a4 },
  })
}
</script>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 19.2px;
}

.form__back {
  align-self: flex-start;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  font-family: var(--font-sans);
  font-size: 12px;
  padding: 0;
}

/* No underline: nothing else in the app has one, and the colour lift is how every other
   text control here says it is live. */
.form__back:hover {
  color: rgb(var(--v-theme-ink));
}
</style>
