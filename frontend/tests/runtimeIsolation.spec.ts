// tests/runtimeIsolation.spec.ts
import { describe, expect, it, vi } from 'vitest'
import { Events } from '@wailsio/runtime'
import { Info } from '~~bindings/github.com/smegg99/s99wails/v2/platform/accent/models.js'

describe('the isolated Wails host', () => {
  it('answers the real runtime bootstrap without a network connection', async () => {
    expect(fetch).toHaveBeenCalledWith('/wails/custom.js', { method: 'HEAD' })
    const response = await fetch('/wails/custom.js', { method: 'HEAD' })
    expect(response.status).toBe(404)
    expect(response.ok).toBe(false)
  })

  it('retains the network boundary when a test restores its globals', async () => {
    const boundary = fetch
    vi.stubGlobal('fetch', vi.fn())
    vi.unstubAllGlobals()
    expect(fetch).toBe(boundary)
    expect((await fetch('/wails/custom.js', { method: 'HEAD' })).status).toBe(404)
  })

  it('keeps generated models and event subscriptions real', () => {
    const info = Info.createFrom({ accent_color: '#123456', dark_mode: true, scheme: 'dark', accent_watch_supported: true })
    expect(info).toBeInstanceOf(Info)
    expect(info.accent_color).toBe('#123456')

    const callback = vi.fn()
    const stop = Events.On('test:isolated-host', callback)
    const host = window as unknown as { _wails: { dispatchWailsEvent: (event: { name: string, data: unknown }) => void } }
    try {
      host._wails.dispatchWailsEvent({ name: 'test:isolated-host', data: info })
      expect(callback).toHaveBeenCalledExactlyOnceWith(expect.objectContaining({ data: info }))
      stop()
      host._wails.dispatchWailsEvent({ name: 'test:isolated-host', data: info })
      expect(callback).toHaveBeenCalledTimes(1)
    }
    finally {
      stop()
    }
  })
})
