// tests/pendingEdits.spec.ts
import { afterEach, describe, expect, it } from 'vitest'
import { flushEdits, registerPendingEdit } from '~/composables/pendingEdits'

const disposers: (() => void)[] = []
afterEach(() => {
  for (const dispose of disposers.splice(0)) dispose()
})

function editor(flush: () => Promise<boolean>) {
  const dispose = registerPendingEdit(flush)
  disposers.push(dispose)
  return dispose
}

describe('pending editor navigation', () => {
  it('coalesces repeated Back while the durable write is outstanding', async () => {
    let finish!: (value: boolean) => void
    let writes = 0
    editor(() => {
      writes++
      return new Promise<boolean>((resolve) => {
        finish = resolve
      })
    })
    const first = flushEdits()
    const second = flushEdits()
    expect(second).toBe(first)
    expect(writes).toBe(1)
    finish(true)
    expect(await first).toBe(true)
  })

  it('stops navigation on refusal and retries the same editor', async () => {
    const calls: string[] = []
    let valid = false
    editor(async () => {
      calls.push('draft')
      return valid
    })
    editor(async () => {
      calls.push('next')
      return true
    })
    expect(await flushEdits()).toBe(false)
    expect(calls).toEqual(['draft'])
    valid = true
    expect(await flushEdits()).toBe(true)
    expect(calls).toEqual(['draft', 'draft', 'next'])
  })

  it('does not write an editor after its owner is disposed', async () => {
    let writes = 0
    const dispose = editor(async () => {
      writes++
      return true
    })
    dispose()
    expect(await flushEdits()).toBe(true)
    expect(writes).toBe(0)
  })
})
