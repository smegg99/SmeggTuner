// tests/web/notice.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { markNoticeSeen, noticeSeen } from '~/web/notice'

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

describe('first-open notice', () => {
  it('shows once and is remembered', () => {
    expect(noticeSeen()).toBe(false)
    markNoticeSeen()
    expect(noticeSeen()).toBe(true)
  })

  it('shows every time when storage throws', () => {
    vi.stubGlobal('localStorage', blockedStorage())
    markNoticeSeen()
    expect(noticeSeen()).toBe(false)
  })
})
