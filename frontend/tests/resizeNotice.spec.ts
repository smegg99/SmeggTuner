import { afterEach, describe, expect, it, vi } from 'vitest'
import { filterResizeNotice } from '~/utils/resizeNotice'

// The log bridge listens for window errors after this filter; the ResizeObserver notice must not reach it.
const disposers: (() => void)[] = []
afterEach(() => disposers.splice(0).forEach(d => d()))

function raise(message: string) {
  window.dispatchEvent(new ErrorEvent('error', { message, cancelable: true }))
}

describe('the ResizeObserver notice', () => {
  it('stays out of the error log and is reported once', () => {
    const once = vi.fn()
    const bridge = vi.fn()
    disposers.push(filterResizeNotice(once))
    window.addEventListener('error', bridge)
    disposers.push(() => window.removeEventListener('error', bridge))

    raise('ResizeObserver loop completed with undelivered notifications.')
    raise('ResizeObserver loop completed with undelivered notifications.')
    expect(bridge).not.toHaveBeenCalled()
    expect(once).toHaveBeenCalledOnce()

    raise('TypeError: x is undefined')
    expect(bridge).toHaveBeenCalledOnce()
  })
})
