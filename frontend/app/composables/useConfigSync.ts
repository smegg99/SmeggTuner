import { effectScope, nextTick, reactive, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import * as ConfigService from '~~bindings/smegg.me/smeggtuner/services/config/service.js'
import { Config } from '~~bindings/smegg.me/smeggtuner/common/config/models.js'

const EVENT_CONFIG_CHANGED = 'config:changed'
const SAVE_DEBOUNCE_MS = 250

const config = reactive(new Config())
let started = false
let ready = false
let applyingBackendConfig = false
let saveTimer: ReturnType<typeof setTimeout> | null = null
let lastSyncedJson = ''

const scope = effectScope(true)

function serialize(source: Partial<Config> = config) {
  return JSON.stringify(new Config(source))
}

function clearSaveTimer() {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
}

function applyFromBackend(source: Partial<Config>) {
  const next = Config.createFrom(source)
  const nextJson = serialize(next)

  applyingBackendConfig = true
  Object.assign(config, next)
  lastSyncedJson = nextJson
  ready = true

  nextTick(() => {
    applyingBackendConfig = false
  })
}

async function saveConfigNow(log: ReturnType<typeof useLogger>) {
  if (!ready || applyingBackendConfig) return

  clearSaveTimer()

  const json = serialize()
  if (json === lastSyncedJson) return

  try {
    await ConfigService.SetConfig(Config.createFrom(JSON.parse(json)))
    lastSyncedJson = json
  }
  catch (err) {
    log.error('config: failed to save', { error: String(err) })
  }
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
    })
  })

  Events.On(EVENT_CONFIG_CHANGED, (ev: { data: unknown }) => {
    const incoming = Config.createFrom(ev.data)
    const incomingJson = serialize(incoming)
    const localJson = serialize()

    if (incomingJson === localJson) {
      lastSyncedJson = incomingJson
      ready = true
      return
    }

    if (ready && localJson !== lastSyncedJson) {
      return
    }

    applyFromBackend(incoming)
  })

  ConfigService.GetConfig()
    .then((c) => {
      applyFromBackend(c)
      log.debug('config: loaded from backend')
    })
    .catch((err) => {
      log.error('config: failed to load', { error: String(err) })
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
