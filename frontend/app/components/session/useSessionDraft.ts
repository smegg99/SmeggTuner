// The new-session dialog's state: the accordion picked or described, the offered name, and the submit.
import { computed, reactive, ref, watch, type Ref } from 'vue'
import { sessionNameFor } from '~/utils/accordion'
import { A4_DEFAULT, A4_MAX, A4_MIN, A4_STEP } from '~/utils/reference'
import type { SessionDraft } from '~/types/session'
import { keyOf } from '~/composables/sessionStore'

export function useSessionDraft(open: Ref<boolean>, saving: () => boolean, submitted: (draft: SessionDraft) => void) {
  const { d } = useI18n()
  const library = useInstruments()

  // Picking one already tuned, or describing the one on the bench. Never both.
  const mode = ref<'pick' | 'new'>('pick')
  const templateId = ref('')
  const acc = reactive({ name: '', a4: A4_DEFAULT })
  const form = reactive({ name: '' })
  const typedName = ref(false)
  const attempted = ref(false)
  const nameError = computed(() => attempted.value && mode.value === 'new' && !acc.name.trim())
  const error = ref('')
  const openedAt = ref(new Date())

  const { draft: a4Draft, commit: commitA4 } = useDraftNumber(
    () => acc.a4,
    (hz) => { acc.a4 = hz },
    { min: A4_MIN, max: A4_MAX, step: A4_STEP },
  )

  // The name follows whichever accordion is in hand, until it is typed over.
  const offeredName = computed(() => sessionNameFor(
    mode.value === 'new' ? acc.name : (library.find(templateId.value)?.name ?? ''),
    d(openedAt.value, 'short'),
  ))

  const submitting = ref(false)
  const busy = computed(() => submitting.value || saving())

  watch(offeredName, (next) => {
    if (!typedName.value) form.name = next
  }, { immediate: true })

  // Reset on every open so a cancelled draft never leaks into the next session.
  watch(open, async (isOpen) => {
    if (!isOpen) return

    mode.value = 'pick'
    templateId.value = ''
    typedName.value = false
    attempted.value = false
    error.value = ''
    openedAt.value = new Date()
    acc.name = ''
    acc.a4 = A4_DEFAULT
    a4Draft.value = String(A4_DEFAULT)
    form.name = offeredName.value

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
    if (busy.value) return
    attempted.value = true
    if (nameError.value) return
    submitting.value = true
    try {
      commitA4()

      const picked = mode.value === 'new' ? await remember() : library.find(templateId.value)
      if (mode.value === 'new' && !picked) return

      error.value = ''
      submitted({
        name: form.name.trim() || offeredName.value,
        notes: '',
        instrument: picked
          ? { ...picked.instrument, name: picked.name }
          : { serial: '', reedCount: 1, banks: [] },
        instrumentId: picked?.id ?? '',
      })
    }
    catch (failure) {
      error.value = keyOf(failure)
    }
    finally {
      submitting.value = false
    }
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
      instrument: { name, serial: '', reedCount: 1, banks: [], a4: acc.a4 },
    })
  }

  return {
    library, mode, templateId, acc, form, nameError, error, a4Draft, commitA4, offeredName, busy,
    typeName, loadFromFile, forget, submit,
  }
}
