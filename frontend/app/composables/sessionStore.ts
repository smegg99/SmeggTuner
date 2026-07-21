// Leaf module: it never imports the sessions composable.
import { ref, shallowRef } from 'vue'
import { Events } from '@wailsio/runtime'
import type { ActiveDTO, ErrorDTO, SessionDTO } from '~~bindings/smegg.me/smeggtuner/services/session/models.js'
import type { Summary } from '~~bindings/smegg.me/smeggtuner/core/session/models.js'

const EVENT_ACTIVE = 'session:active'
const EVENT_SAVE_FAILED = 'session:saveFailed'

// Every backend error is an i18n key (services/session/events.go); anything else is a bug.
const ERROR_KEY_PATTERN = /(?:session|record)\.error\.[A-Za-z0-9]+/
const ERROR_UNEXPECTED = 'session.error.unexpected'

export const list = shallowRef<Summary[]>([])
export const active = shallowRef<SessionDTO | null>(null)
export const loading = ref(false)
export const busy = ref(false)
export const error = ref('')

export function keyOf(err: unknown): string {
  const text = err instanceof Error ? err.message : String(err)
  return ERROR_KEY_PATTERN.exec(text)?.[0] ?? ERROR_UNEXPECTED
}

// One busy flag, one error key. Returns undefined on failure: the caller checks it or reads `error`.
export async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
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

let subscribed = false

// The service is the source of truth for what is open: Create/Open/Close publish.
export function ensureSubscribed() {
  if (subscribed) return
  subscribed = true
  Events.On(EVENT_ACTIVE, (e: { data: ActiveDTO }) => {
    active.value = e.data?.session ?? null
  })
  Events.On(EVENT_SAVE_FAILED, (e: { data: ErrorDTO }) => {
    error.value = e.data?.key ?? ERROR_UNEXPECTED
  })
}
