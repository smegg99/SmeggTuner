import { computed, ref, shallowRef } from 'vue'
import { Events } from '@wailsio/runtime'
import * as RecordService from '~~bindings/smegg.me/smeggtuner/services/record/service.js'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import type { StateDTO, TableDTO } from '~~bindings/smegg.me/smeggtuner/services/record/models.js'

// Holds the tuning table the backend pushes for the open session, plus the technician's edits to it.

const EVENT_STATE = 'record:state'
const EVENT_TABLE = 'record:table'

// A reading with no session open; the UI offers to create one rather than erroring.
const ERROR_NO_SESSION = 'session.error.noSession'
const ERROR_KEY_PATTERN = /(?:session|record)\.error\.[A-Za-z0-9]+/
const ERROR_UNEXPECTED = 'record.error.unexpected'

const sessionId = ref('')
const readings = ref(0)
const armed = ref(false)
const table = shallowRef<TableDTO | null>(null)
const busy = ref(false)
const error = ref('')

let subscribed = false

function keyOf(err: unknown): string {
  const text = err instanceof Error ? err.message : String(err)
  return ERROR_KEY_PATTERN.exec(text)?.[0] ?? ERROR_UNEXPECTED
}

function applyState(s: StateDTO | null | undefined) {
  sessionId.value = s?.sessionId ?? ''
  readings.value = s?.readings ?? 0
  armed.value = s?.armed ?? false
}

async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
  busy.value = true
  error.value = ''
  try {
    return await fn()
  }
  catch (err) {
    error.value = keyOf(err)
    return undefined
  }
  finally {
    busy.value = false
  }
}

export function useRecord() {
  if (!subscribed) {
    subscribed = true

    // Backend-driven: every locked note upserts a reading in Go and the table arrives already scored.
    Events.On(EVENT_STATE, (e: { data: StateDTO }) => applyState(e.data))
    Events.On(EVENT_TABLE, (e: { data: TableDTO }) => {
      table.value = e.data ?? null
    })
  }

  async function load() {
    applyState(await RecordService.State())
    table.value = await run(() => RecordService.Table()) ?? null
  }

  async function show() {
    const t = await run(() => RecordService.Table())
    if (t) table.value = t
  }

  async function undo() {
    const t = await run(() => RecordService.Undo())
    if (t) table.value = t
  }

  // The one control over whether a lock is saved; state comes from the backend, never set optimistically.
  async function setArmed(on: boolean) {
    await run(() => RecordService.SetArmed(on))
  }

  // Both entry points (R shortcut, toolbar) route here; arming makes the engine re-enter its own warm-up, which the note card reads (no UI timer).
  function toggleRecording(on: boolean) {
    void setArmed(on)
  }

  async function deleteTake(take: number) {
    const done = await run(() => SessionService.DeleteTake(take))
    if (done !== undefined) await show()
  }

  async function clear() {
    const t = await run(() => RecordService.Clear())
    if (t) table.value = t
  }

  // The backend recomputes the row against the curve; the UI never derives a tuning number.
  async function editReed(take: number, reed: number, value: number, unit: string) {
    const t = await run(() => RecordService.EditReed(take, reed, value, unit))
    if (t) table.value = t
  }

  return {
    sessionId,
    readings,
    armed,
    table,
    busy,
    error,
    // Armed, not merely open: a session opens in warm-up and saves nothing until the record key.
    recording: computed(() => sessionId.value !== '' && armed.value),
    // The UI's cue to offer creating a session inline.
    needsSession: computed(() => error.value === ERROR_NO_SESSION),
    load,
    show,
    undo,
    setArmed,
    toggleRecording,
    deleteTake,
    clear,
    editReed,
  }
}
