// tests/web/storage.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { loadConfig, saveConfig } from '~/web/storage'

afterEach(() => {
  vi.unstubAllGlobals()
  localStorage.clear()
})

// Private mode and blocked site data: every access throws.
function blockedStorage() {
  const fail = () => {
    throw new DOMException('blocked', 'SecurityError')
  }
  return { getItem: fail, setItem: fail, removeItem: fail, clear: fail, key: fail, length: 0 }
}

describe('browser config storage', () => {
  it('round-trips the config', () => {
    saveConfig('{"tuner":{"a4":442}}')
    expect(loadConfig()).toBe('{"tuner":{"a4":442}}')
  })

  it('loadConfig returns empty when localStorage throws', () => {
    vi.stubGlobal('localStorage', blockedStorage())
    expect(loadConfig()).toBe('')
  })

  it('saveConfig does not throw when localStorage throws', () => {
    vi.stubGlobal('localStorage', blockedStorage())
    expect(() => saveConfig('{}')).not.toThrow()
  })
})
