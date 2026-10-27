// @vitest-environment node
// tests/configSync.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { Config, Device } from '~~bindings/smegg.me/smeggtuner/backend/common/config/models.js'

const logError = vi.fn()
let server: Config

const clone = (value: Config) => Config.createFrom(JSON.stringify(value))
const response = (theme: string) => {
  const value = clone(server)
  value.preferences.theme = theme
  return value
}

const backend = vi.hoisted(() => ({
  get: vi.fn(),
  patch: vi.fn(),
  changed: undefined as ((event: { data: Config }) => void) | undefined,
}))

vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/config/service.js', () => ({
  GetConfig: backend.get,
  PatchConfig: backend.patch,
}))

vi.mock('@wailsio/runtime', async original => ({
  ...await original<typeof import('@wailsio/runtime')>(),
  Events: { On: (_name: string, callback: typeof backend.changed) => { backend.changed = callback } },
}))

function pending() {
  let resolve!: (config: Config) => void
  let reject!: (error: Error) => void
  const promise = new Promise<Config>((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}

async function flush() {
  await Promise.resolve()
  await nextTick()
  await Promise.resolve()
}

async function start() {
  const { useConfigSync } = await import('~/composables/useConfigSync')
  const sync = useConfigSync()
  await flush()
  await flush()
  return sync
}

beforeEach(() => {
  vi.resetModules()
  vi.useFakeTimers()
  backend.get.mockReset()
  backend.patch.mockReset()
  backend.changed = undefined
  logError.mockReset()
  server = new Config()
  server.preferences.theme = 'auto'
  server.tuner.a4 = 440
  server.audio.devices.mic = new Device({ calibration: 'uncalibrated', calibration_percent: 0 })
  backend.get.mockImplementation(async () => clone(server))
  backend.patch.mockImplementation((changes: Partial<Config>) => {
    if (changes.preferences) Object.assign(server.preferences, changes.preferences)
    if (changes.tuner) Object.assign(server.tuner, changes.tuner)
    return clone(server)
  })
  vi.stubGlobal('useLogger', () => ({ debug: vi.fn(), error: logError }))
})

afterEach(() => {
  vi.clearAllTimers()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('config autosave ordering', () => {
  it('keeps only one write active and waits for the latest edit when explicitly flushed', async () => {
    const first = pending()
    const second = pending()
    backend.patch.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    await vi.advanceTimersByTimeAsync(250)
    sync.config.preferences.theme = 'dark'
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledTimes(1)

    let saved = false
    const saving = sync.saveConfigNow().then(() => {
      saved = true
    })
    backend.changed!({ data: response('light') })
    server = response('light')
    first.resolve(clone(server))
    await flush()
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(backend.patch.mock.calls[1]![0]).toEqual({ preferences: { theme: 'dark' } })
    expect(sync.config.preferences.theme).toBe('dark')
    expect(saved).toBe(false)

    backend.changed!({ data: response('dark') })
    server = response('dark')
    second.resolve(clone(server))
    await saving
    await vi.advanceTimersByTimeAsync(250)
    expect(saved).toBe(true)
    expect(backend.patch).toHaveBeenCalledTimes(2)
  })

  it('persists a revert to the saved value instead of adopting the older in-flight echo', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    sync.config.preferences.theme = 'auto'
    await flush()
    backend.changed!({ data: response('light') })
    server = response('light')
    first.resolve(clone(server))
    await saving
    expect(sync.config.preferences.theme).toBe('auto')
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(backend.patch.mock.calls[1]![0]).toEqual({ preferences: { theme: 'auto' } })
  })

  it('coalesces edits during a write into the latest snapshot', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    sync.config.preferences.theme = 'dark'
    sync.config.preferences.language = 'pl'
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledTimes(1)
    first.resolve(response('light'))
    await saving
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(backend.patch.mock.calls[1]![0].preferences).toMatchObject({ theme: 'dark', language: 'pl' })
  })

  it.each([false, true])('drains a newer edit after rejection without retrying it indefinitely (newer rejects: %s)', async (newerRejects) => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    if (newerRejects) backend.patch.mockRejectedValueOnce(new Error('newer write failed'))
    const sync = await start()
    sync.config.preferences.theme = 'light'
    await vi.advanceTimersByTimeAsync(250)
    sync.config.preferences.theme = 'dark'
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(0)

    const saving = sync.saveConfigNow()
    first.reject(new Error('first write failed'))
    await saving
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(backend.patch.mock.calls[1]![0]).toEqual({ preferences: { theme: 'dark' } })
    expect(logError).toHaveBeenCalledTimes(newerRejects ? 2 : 1)
    expect(logError).toHaveBeenCalledWith('log.frontendConfigSaveFailed', { error: 'Error: first write failed' })
    await vi.advanceTimersByTimeAsync(1000)
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(vi.getTimerCount()).toBe(0)
  })

  it('releases the write guard after failure so the dirty value can be retried', async () => {
    backend.patch.mockRejectedValueOnce(new Error('write failed'))
    const sync = await start()
    sync.config.preferences.theme = 'dark'
    await sync.saveConfigNow()
    await sync.saveConfigNow()
    expect(backend.patch).toHaveBeenCalledTimes(2)
    expect(backend.patch.mock.calls[1]![0]).toEqual({ preferences: { theme: 'dark' } })
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledTimes(2)
  })

  it('adopts clean backend updates without writing them back', async () => {
    const sync = await start()
    const changed = Config.createFrom(JSON.stringify(sync.config))
    changed.preferences.theme = 'dark'
    backend.changed!({ data: changed })
    await vi.advanceTimersByTimeAsync(250)
    expect(sync.config.preferences.theme).toBe('dark')
    expect(backend.patch).not.toHaveBeenCalled()
  })

  it('merges calibration and unrelated preference events while preserving dirty leaves', async () => {
    const sync = await start()
    sync.config.preferences.theme = 'light'
    server.preferences.theme = 'dark'
    server.preferences.language = 'pl'
    server.audio.devices.mic!.calibration = 'calibrated'
    server.audio.devices.mic!.calibration_percent = 0.125
    backend.changed!({ data: clone(server) })
    expect(sync.config.preferences).toMatchObject({ theme: 'light', language: 'pl' })
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.125)
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledExactlyOnceWith({ preferences: { theme: 'light' } })
    expect(server.audio.devices.mic!.calibration_percent).toBe(0.125)
  })

  it('does not submit stale A4 or calibration with a later preference change', async () => {
    const sync = await start()
    server.tuner.a4 = 442
    server.audio.devices.mic!.calibration_percent = 0.125
    expect(sync.config.tuner.a4).toBe(440)
    sync.config.preferences.theme = 'light'
    await sync.saveConfigNow()
    expect(backend.patch).toHaveBeenCalledExactlyOnceWith({ preferences: { theme: 'light' } })
    expect(sync.config.tuner.a4).toBe(442)
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.125)
    sync.config.preferences.language = 'pl'
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenLastCalledWith({ preferences: { language: 'pl' } })
    expect(server.tuner.a4).toBe(442)
  })

  it('patches only the edited leaf inside a device', async () => {
    const sync = await start()
    sync.config.audio.devices.mic!.auto_volume = true
    server.audio.devices.mic!.calibration_percent = 0.125
    backend.patch.mockImplementationOnce(() => {
      server.audio.devices.mic!.auto_volume = true
      return clone(server)
    })
    await sync.saveConfigNow()
    expect(backend.patch).toHaveBeenCalledExactlyOnceWith({ audio: { devices: { mic: { auto_volume: true } } } })
    expect(sync.config.audio.devices.mic).toMatchObject({ auto_volume: true, calibration_percent: 0.125 })
  })

  it('accepts normalized submitted leaves and rebases only subsequent edits', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.tuner.tolerance = 1.234
    const saving = sync.saveConfigNow()
    sync.config.preferences.language = 'pl'
    server.tuner.tolerance = 1.2
    server.tuner.a4 = 442
    first.resolve(clone(server))
    await saving
    expect(backend.patch.mock.calls.map(([changes]) => changes)).toEqual([
      { tuner: { tolerance: 1.234 } },
      { preferences: { language: 'pl' } },
    ])
    expect(sync.config.tuner).toMatchObject({ tolerance: 1.2, a4: 442 })
  })

  it('keeps a local revert through in-flight calibration events and the response', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    sync.config.preferences.theme = 'auto'
    server.preferences.theme = 'light'
    server.audio.devices.mic!.calibration_percent = 0.125
    backend.changed!({ data: clone(server) })
    expect(sync.config.preferences.theme).toBe('auto')
    first.resolve(clone(server))
    await saving
    expect(backend.patch).toHaveBeenLastCalledWith({ preferences: { theme: 'auto' } })
    expect(sync.config.preferences.theme).toBe('auto')
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.125)
  })

  it.each([[2, 3], []])('replaces an array as one leaf: %j', async (...boxes) => {
    server.tuner.reed_boxes = [1]
    backend.get.mockResolvedValue(clone(server))
    const sync = await start()
    sync.config.tuner.reed_boxes = boxes
    await sync.saveConfigNow()
    expect(backend.patch).toHaveBeenCalledExactlyOnceWith({ tuner: { reed_boxes: boxes } })
    sync.config.tuner.reed_boxes = [...boxes]
    await vi.advanceTimersByTimeAsync(250)
    expect(backend.patch).toHaveBeenCalledTimes(1)
  })

  it('merges the latest deferred event after failure without retrying an unchanged edit', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    server.tuner.a4 = 441
    backend.changed!({ data: clone(server) })
    server.tuner.a4 = 442
    backend.changed!({ data: clone(server) })
    first.reject(new Error('write failed'))
    await saving
    expect(sync.config.tuner.a4).toBe(442)
    expect(sync.config.preferences.theme).toBe('light')
    await vi.advanceTimersByTimeAsync(1000)
    expect(backend.patch).toHaveBeenCalledTimes(1)
    await sync.saveConfigNow()
    expect(backend.patch).toHaveBeenLastCalledWith({ preferences: { theme: 'light' } })
  })

  it.each([false, true])('refreshes newer calibration after an older patch reply (local revert: %s)', async (revert) => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    server.preferences.theme = 'light'
    const olderReply = clone(server)
    if (revert) sync.config.preferences.theme = 'auto'
    server.audio.devices.mic!.calibration_percent = 0.125
    backend.changed!({ data: clone(server) })
    first.resolve(olderReply)
    await saving
    expect(backend.get).toHaveBeenCalledTimes(2)
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.125)
    expect(sync.config.preferences.theme).toBe(revert ? 'auto' : 'light')
    expect(backend.patch).toHaveBeenCalledTimes(revert ? 2 : 1)
    if (revert) expect(backend.patch).toHaveBeenLastCalledWith({ preferences: { theme: 'auto' } })
  })

  it('refreshes again for events received during the read and preserves edits made during it', async () => {
    const first = pending()
    const read = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    backend.get.mockReturnValueOnce(read.promise)
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    server.preferences.theme = 'light'
    backend.changed!({ data: clone(server) })
    first.resolve(clone(server))
    await flush()
    expect(backend.get).toHaveBeenCalledTimes(2)
    const olderRead = clone(server)
    sync.config.preferences.language = 'pl'
    server.audio.devices.mic!.calibration_percent = 0.25
    backend.changed!({ data: clone(server) })
    read.resolve(olderRead)
    await saving
    expect(backend.get).toHaveBeenCalledTimes(3)
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.25)
    expect(sync.config.preferences.language).toBe('pl')
    expect(backend.patch).toHaveBeenLastCalledWith({ preferences: { language: 'pl' } })
  })

  it('logs a failed refresh without spinning and retains it for the next flush', async () => {
    const first = pending()
    backend.patch.mockReturnValueOnce(first.promise)
    const sync = await start()
    backend.get.mockRejectedValueOnce(new Error('refresh failed'))
    sync.config.preferences.theme = 'light'
    const saving = sync.saveConfigNow()
    server.preferences.theme = 'light'
    const olderReply = clone(server)
    server.audio.devices.mic!.calibration_percent = 0.125
    backend.changed!({ data: clone(server) })
    first.resolve(olderReply)
    await saving
    await vi.advanceTimersByTimeAsync(1000)
    expect(backend.get).toHaveBeenCalledTimes(2)
    expect(backend.patch).toHaveBeenCalledTimes(1)
    expect(logError).toHaveBeenCalledWith('log.frontendConfigLoadFailed', { error: 'Error: refresh failed' })
    await sync.saveConfigNow()
    expect(backend.get).toHaveBeenCalledTimes(3)
    expect(sync.config.audio.devices.mic!.calibration_percent).toBe(0.125)
  })
})
