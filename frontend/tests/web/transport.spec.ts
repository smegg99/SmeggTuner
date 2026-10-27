// tests/web/transport.spec.ts
import { describe, expect, it } from 'vitest'
import { createWorkerTransport } from '~/web/transport'
import { fakePort } from './fakePort'

const callArgs = (methodID: number, args: unknown[]) => ({ 'call-id': 'x', methodID, args })

describe('worker transport', () => {
  it('posts a call and resolves with data', async () => {
    const { port, posted, reply } = fakePort()
    const pending = createWorkerTransport(port).call(0, 0, '', callArgs(42, [1, 'a']))
    expect(posted).toEqual([{ type: 'call', id: 1, methodID: 42, args: '[1,"a"]' }])
    reply({ type: 'result', id: 1, reply: '{"ok":true,"data":5}' })
    await expect(pending).resolves.toBe(5)
  })

  it('rejects with RuntimeError carrying the cause', async () => {
    const { port, reply } = fakePort()
    const pending = createWorkerTransport(port).call(0, 0, '', callArgs(7, []))
    reply({ type: 'result', id: 1, reply: '{"ok":false,"kind":"RuntimeError","message":"m","cause":{"key":"k"}}' })
    const err = await pending.catch((e: Error & { cause?: unknown }) => e)
    expect(err.name).toBe('RuntimeError')
    expect(err.message).toBe('m')
    expect(err.cause).toEqual({ key: 'k' })
  })

  it('maps TypeError and ReferenceError kinds to their classes', async () => {
    const { port, reply } = fakePort()
    const t = createWorkerTransport(port)
    const a = t.call(0, 0, '', callArgs(1, []))
    const b = t.call(0, 0, '', callArgs(2, []))
    reply({ type: 'result', id: 1, reply: '{"ok":false,"kind":"TypeError","message":"t"}' })
    reply({ type: 'result', id: 2, reply: '{"ok":false,"kind":"ReferenceError","message":"r"}' })
    await expect(a).rejects.toBeInstanceOf(TypeError)
    await expect(b).rejects.toBeInstanceOf(ReferenceError)
  })

  it('resolves non-call objects to undefined without posting', async () => {
    const { port, posted } = fakePort()
    await expect(createWorkerTransport(port).call(6, 3, '', {})).resolves.toBeUndefined()
    expect(posted).toEqual([])
  })
})

describe('worker transport with page-side methods', () => {
  it('answers a local method without posting', async () => {
    const { port, posted } = fakePort()
    const t = createWorkerTransport(port, new Map([[9, (args: unknown[]) => `got ${args[0]}`]]))
    await expect(t.call(0, 0, '', callArgs(9, ['x']))).resolves.toBe('got x')
    expect(posted).toEqual([])
  })
})
