// The accordion tab's autosave: every committed change to the panel's value is written, a moment after
// the last one, and navigation waits for the last durable write.
import { onBeforeUnmount, ref, watch, type ComputedRef, type Ref } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/backend/services/session/service.js'
import type { InstrumentTemplate } from '~/types/session'
import { registerPendingEdit } from '~/composables/pendingEdits'
import { error, keyOf } from '~/composables/sessionStore'

// Long enough that typing a name is one write, short enough that leaving the tab never races it.
const AUTOSAVE_MS = 600

interface Autosave {
  /** everything the panel may write, as one value; drafts stay out until committed */
  edited: ComputedRef<string>
  /** the accordion the tab shows; loading another one is not an edit */
  sourceID: () => string | undefined
  collect: () => InstrumentTemplate | null
  /** commits the typed drafts before a navigation waits on the write */
  commitDrafts: () => void
  columnsMoved: Readonly<Ref<boolean>>
  nameError: Readonly<Ref<boolean>>
  /** the library entry the accordion was taken from, kept in step with the session */
  libraryID: () => string | undefined
}

export function useAccordionAutosave(o: Autosave) {
  const { active, refresh } = useSessions()
  const library = useInstruments()
  const sessionID = active.value?.id

  // What is already stored. Loading an accordion sets it, so arriving on the tab is not an edit.
  const stored = ref(o.edited.value)
  watch(o.sourceID, () => {
    stored.value = o.edited.value
  }, { flush: 'post' })

  let timer: ReturnType<typeof setTimeout> | undefined
  let disposed = false
  let saving: Promise<boolean> | undefined

  watch(o.edited, (now) => {
    if (now === stored.value) return
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => void save(), AUTOSAVE_MS)
  })

  const removePending = registerPendingEdit(async () => {
    o.commitDrafts()
    if (timer) clearTimeout(timer)
    timer = undefined
    do {
      if (!await save()) return false
    } while (o.edited.value !== stored.value)
    error.value = ''
    return true
  })
  onBeforeUnmount(() => {
    disposed = true
    removePending()
    if (timer) clearTimeout(timer)
  })

  function save(): Promise<boolean> {
    if (saving) return saving
    if (timer) clearTimeout(timer)
    timer = undefined
    if (o.edited.value === stored.value) return Promise.resolve(true)
    saving = (async () => {
      do {
        if (disposed) return false
        if (timer) clearTimeout(timer)
        timer = undefined
        const next = o.collect()
        if (!next || o.columnsMoved.value) {
          error.value = o.columnsMoved.value ? 'session.error.hasReadings' : o.nameError.value ? 'session.error.instrumentName' : 'session.error.invalidInstrument'
          return false
        }
        if (!sessionID || active.value?.id !== sessionID) {
          error.value = 'session.error.noSession'
          return false
        }
        const was = o.edited.value
        const libraryID = o.libraryID()
        try {
          await SessionService.SetInstrument(sessionID, next.instrument as never)
          if (libraryID) await library.save({ ...next, id: libraryID })
          await SessionService.Save()
          await refresh()
          stored.value = was
        }
        catch (failure) {
          error.value = keyOf(failure)
          // Try a newer draft once; an unchanged failed snapshot waits for an explicit retry.
          if (o.edited.value === was) return false
          continue
        }
        if (o.edited.value === stored.value) {
          error.value = ''
          return true
        }
      }
      while (!disposed)
      return false
    })().finally(() => { saving = undefined })
    return saving
  }

  return { save }
}
