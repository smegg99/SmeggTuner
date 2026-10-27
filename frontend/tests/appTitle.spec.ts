import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAppTitle } from '~/composables/s99wails'

/*
 * The generated composable, tested here because s99wails has no JS runner. What
 * matters is how many pushes reach the binding. useAppTitle no-ops off the
 * desktop, hence the marker.
 */
beforeEach(() => {
  ;(window as unknown as Record<string, unknown>)._wails = {}
})

const flush = () => new Promise(resolve => setTimeout(resolve, 0))

describe('the app title composable', () => {
  it('pushes the phrase to the binding', async () => {
    const Set = vi.fn().mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('tuning Bugari 12/41')
    await flush()

    expect(Set).toHaveBeenCalledExactlyOnceWith('tuning Bugari 12/41')
  })

  // A watcher over several sources fires once per source.
  it('coalesces a burst into the last value', async () => {
    const Set = vi.fn().mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('tuning Bugari 10/41')
    setTitle('tuning Bugari 11/41')
    setTitle('tuning Bugari 12/41')
    await flush()

    expect(Set).toHaveBeenCalledExactlyOnceWith('tuning Bugari 12/41')
  })

  it('does not re-push a phrase that has not changed', async () => {
    const Set = vi.fn().mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('workshop')
    await flush()
    setTitle('workshop')
    await flush()

    expect(Set).toHaveBeenCalledTimes(1)
  })

  it('pushes again after a change', async () => {
    const Set = vi.fn().mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('workshop')
    await flush()
    setTitle('settings')
    await flush()

    expect(Set).toHaveBeenNthCalledWith(2, 'settings')
  })

  // A failed push must not poison the cache.
  it('retries the same phrase after a failure', async () => {
    const Set = vi.fn().mockRejectedValueOnce(new Error('binding down')).mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('tuning')
    await flush()
    setTitle('tuning')
    await flush()

    expect(Set).toHaveBeenCalledTimes(2)
  })

  // Regression guard for the bug this rewrite fixed: the old composable marked
  // a push "confirmed" as soon as it was SENT, not once the binding promise
  // settled. That meant an identical setTitle arriving while the first call was
  // still in flight got deduped against a value that had not actually landed
  // yet - so if the in-flight call then FAILED, the dedupe left nothing to
  // retry and the title stuck on whatever was showing before. Here the second,
  // identical push lands mid-flight (before the first Set's promise resolves
  // or rejects), and the first call then rejects. The text must still reach
  // the binding: confirmation has to wait for success, not for send.
  it('still delivers the text after a same-value push races an in-flight failure', async () => {
    let rejectFirst: (err: Error) => void = () => {}
    const firstCall = new Promise<void>((_, reject) => {
      rejectFirst = reject
    })
    const Set = vi.fn().mockReturnValueOnce(firstCall).mockResolvedValue(undefined)
    const { setTitle } = useAppTitle({ Set })

    setTitle('tuning')
    await flush()
    expect(Set).toHaveBeenCalledTimes(1)

    // Same text, pushed while the first call is still unsettled.
    setTitle('tuning')
    await flush()

    // Now let the in-flight call fail.
    rejectFirst(new Error('binding down'))
    await flush()
    await flush()

    expect(Set).toHaveBeenLastCalledWith('tuning')
    expect(Set.mock.calls.length).toBeGreaterThanOrEqual(2)
  })
})
