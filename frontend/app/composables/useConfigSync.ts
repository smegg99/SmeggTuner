// app/composables/useConfigSync.ts
import { effectScope, reactive, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import * as ConfigService from '~~bindings/smegg.me/smeggtuner/backend/services/config/service.js'
import { Config } from '~~bindings/smegg.me/smeggtuner/backend/common/config/models.js'

const EVENT_CONFIG_CHANGED = 'config:changed'
const SAVE_DEBOUNCE_MS = 250

const config = reactive(new Config())
let started = false
let ready = false
let applyingBackendConfig = false
let saveTimer: ReturnType<typeof setTimeout> | null = null
let saving: Promise<void> | null = null
let lastSyncedJson = ''
let needsRefresh = false

const scope = effectScope(true)

function serialize(source: Partial<Config> = config) {
  return JSON.stringify(new Config(source))
}

type Changes = Record<string, unknown>

function isObject(value: unknown): value is Changes {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

// Editors change leaves; arrays are whole values and absent keys do not mean deletion.
function diff(base: Changes, draft: Changes): Changes {
  return Object.fromEntries(Object.entries(draft).flatMap(([key, value]) => {
    const previous = base[key]
    if (isObject(value)) {
      const changes = diff(isObject(previous) ? previous : {}, value)
      return Object.keys(changes).length ? [[key, changes]] : []
    }
    return JSON.stringify(value) === JSON.stringify(previous) ? [] : [[key, value]]
  }))
}

function merge(base: Changes, changes: Changes): Changes {
  return {
    ...base,
    ...Object.fromEntries(Object.entries(changes).map(([key, value]) => [
      key,
      isObject(value) ? merge(isObject(base[key]) ? base[key] : {}, value) : value,
    ])),
  }
}

function clearSaveTimer() {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
}

function applyFromBackend(source: Partial<Config>, changes: Changes = {}) {
  lastSyncedJson = serialize(source)
  applyingBackendConfig = true
  Object.assign(config, Config.createFrom(merge(JSON.parse(lastSyncedJson), changes)))
  applyingBackendConfig = false
  ready = true
}

function mergeFromBackend(source: Partial<Config>) {
  const changes = ready ? diff(JSON.parse(lastSyncedJson), JSON.parse(serialize())) : {}
  applyFromBackend(source, changes)
}

async function refreshDeferred(log: ReturnType<typeof useLogger>) {
  while (needsRefresh) {
    needsRefresh = false
    try {
      mergeFromBackend(await ConfigService.GetConfig())
    }
    catch (err) {
      needsRefresh = true
      log.error('log.frontendConfigLoadFailed', { error: String(err) })
      return
    }
  }
}

async function saveConfigNow(log: ReturnType<typeof useLogger>) {
  if (!ready || applyingBackendConfig) return

  clearSaveTimer()
  if (saving) return saving

  // Only one snapshot may reach the backend at a time; drain edits made while it writes.
  saving = (async () => {
    while (true) {
      clearSaveTimer()
      // Events and replies have no shared revision; read after deferred events before draining edits.
      if (needsRefresh) await refreshDeferred(log)
      const json = serialize()
      const submitted = JSON.parse(json)
      const changes = diff(JSON.parse(lastSyncedJson), submitted)
      if (!Object.keys(changes).length) return

      try {
        const updated = await ConfigService.PatchConfig(changes)
        // Only post-submission edits survive normalization, including a revert to the old baseline.
        applyFromBackend(updated, diff(submitted, JSON.parse(serialize())))
      }
      catch (err) {
        log.error('log.frontendConfigSaveFailed', { error: String(err) })
        if (needsRefresh) await refreshDeferred(log)
        const remaining = diff(JSON.parse(lastSyncedJson), JSON.parse(serialize()))
        // A newer edit may have spent its debounce waiting here; drain it, but never spin on a failed payload.
        if (JSON.stringify(remaining) === JSON.stringify(changes)) return
      }
    }
  })().finally(() => { saving = null })
  return saving
}

function scheduleSave(log: ReturnType<typeof useLogger>) {
  if (!ready || applyingBackendConfig) return

  clearSaveTimer()
  saveTimer = setTimeout(() => {
    saveTimer = null
    void saveConfigNow(log)
  }, SAVE_DEBOUNCE_MS)
}

function start(log: ReturnType<typeof useLogger>) {
  if (started) return
  started = true

  scope.run(() => {
    watch(() => JSON.stringify(config), (json) => {
      if (json === lastSyncedJson) return
      scheduleSave(log)
    }, { flush: 'sync' })
  })

  Events.On(EVENT_CONFIG_CHANGED, (ev: { data: unknown }) => {
    if (saving) {
      needsRefresh = true
      return
    }
    mergeFromBackend(Config.createFrom(ev.data))
  })

  ConfigService.GetConfig()
    .then((c) => {
      applyFromBackend(c)
      log.debug('log.frontendConfigLoaded')
    })
    .catch((err) => {
      log.error('log.frontendConfigLoadFailed', { error: String(err) })
    })
}

export function useConfigSync() {
  const log = useLogger()
  start(log)

  return {
    config,
    saveConfigNow: () => saveConfigNow(log),
  }
}
