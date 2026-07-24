import { computed } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import { active, busy, ensureSubscribed, error, keyOf, list, loading, run } from '~/composables/sessionStore'
import { useSessionCurve } from '~/composables/sessionCurve'
import type { SessionDTO } from '~~bindings/smegg.me/smeggtuner/services/session/models.js'
import type { SessionDraft } from '~/components/session/SessionDialog.vue'

export function useSessions() {
  // Native file dialogs have no language of their own: Go is handed the words.
  const { t } = useI18n()

  ensureSubscribed()

  async function refresh() {
    loading.value = true
    error.value = ''
    try {
      list.value = await SessionService.List()
    }
    catch (err) {
      error.value = keyOf(err)
    }
    finally {
      loading.value = false
    }
  }

  async function load() {
    active.value = await SessionService.Active()
    await refresh()
  }

  // Create also opens the session (inline creation on Record); the draft carries the whole instrument.
  async function create(draft: SessionDraft): Promise<SessionDTO | undefined> {
    const created = await run(() => SessionService.Create({
      name: draft.name,
      notes: draft.notes,
      instrument: draft.instrument,
      instrumentId: draft.instrumentId,
    }))
    if (created) await refresh()
    return created ?? undefined
  }

  // Sets the register; a register has as many banks as reeds.
  async function setRegister(name: string) {
    return run(() => SessionService.SetRegister(name))
  }

  // Turns the bench toward the bass keyboard, or back to the treble.
  async function setBass(on: boolean) {
    return run(() => SessionService.SetBass(on))
  }

  // Pulls a bass switch; empty is the whole (or fixed) machine.
  async function setBassRegister(name: string) {
    return run(() => SessionService.SetBassRegister(name))
  }

  // Import always creates a new session, never overwriting the open one.
  async function importFile() {
    const path = await run(() => SessionService.OpenFileDialog(
      t('session.file.openTitle'),
      t('session.file.openFilter'),
    ))
    if (!path) return undefined

    const added = await run(() => SessionService.ImportSession(path))
    if (added) await refresh()
    return added ?? undefined
  }

  // Write one out whole; the backend suggests the .stsf file name.
  async function exportFile(id: string, name: string) {
    const suggested = await SessionService.SuggestSessionFileName(name)
    const path = await run(() => SessionService.SaveFileDialog(
      'session',
      suggested,
      t('session.file.exportSessionTitle'),
      t('session.file.openFilter'),
    ))
    if (!path) return false

    return (await run(() => SessionService.ExportSession(id, path))) !== undefined
  }

  async function open(id: string) {
    return run(() => SessionService.Open(id))
  }

  async function close() {
    return run(() => SessionService.Close())
  }

  async function remove(id: string) {
    const done = await run(() => SessionService.Delete(id))
    if (done !== undefined) await refresh()
  }

  // Refused while a pass is open: a pass measured against a moving reference is worthless.
  async function setA4(hz: number) {
    return run(() => SessionService.SetA4(hz))
  }

  // Every mutation already schedules a write; this is the explicit flush.
  async function save() {
    return run(() => SessionService.Save())
  }

  return {
    list,
    active,
    loading,
    busy,
    error,
    readings: computed(() => active.value?.readings ?? 0),
    refresh,
    load,
    create,
    open,
    close,
    remove,
    setA4,
    save,
    setRegister,
    setBass,
    setBassRegister,
    importFile,
    exportFile,
    ...useSessionCurve(),
  }
}
