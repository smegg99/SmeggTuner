// tests/accordionAutosave.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, shallowRef } from 'vue'
import AccordionPanel from '~/components/session/AccordionPanel.vue'
import { flushEdits } from '~/composables/pendingEdits'
import { error } from '~/composables/sessionStore'
import type { Instrument, InstrumentTemplate } from '~/types/session'

const backend = vi.hoisted(() => ({ set: vi.fn(), save: vi.fn() }))
vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/session/service.js', () => ({
  SetInstrument: backend.set,
  Save: backend.save,
}))

const source = shallowRef<InstrumentTemplate | null>(null)
vi.mock('~/composables/useSessionAccordion', () => ({
  useSessionAccordion: () => ({ source, remembered: shallowRef(false) }),
}))

let wrapper: ReturnType<typeof mount>
let saved: Instrument
const active = shallowRef({ id: 'session', readings: 0, instrument: {} as Instrument })

function deferred() {
  let resolve!: () => void
  let reject!: (error: Error) => void
  const promise = new Promise<void>((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}

beforeEach(() => {
  vi.useFakeTimers()
  backend.set.mockReset()
  backend.save.mockReset().mockResolvedValue(undefined)
  saved = { name: 'Original', serial: '', reedCount: 1, banks: [], a4: 440.5 }
  source.value = { id: '', name: 'Original', hasImage: false, instrument: saved }
  active.value = { id: 'session', readings: 0, instrument: saved }
  error.value = ''
  backend.set.mockImplementation(async (_id: string, instrument: Instrument) => {
    saved = instrument
  })
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  vi.stubGlobal('useSessions', () => ({
    active,
    refresh: async () => {
      active.value = { ...active.value, instrument: saved }
      source.value = { ...source.value!, name: saved.name, instrument: saved }
    },
  }))
  vi.stubGlobal('useInstruments', () => ({ save: vi.fn() }))
  wrapper = mount(AccordionPanel, {
    shallow: true,
    global: {
      stubs: {
        'InstrumentIdentityFields': {
          props: ['name'],
          emits: ['update:name', 'submit'],
          template: '<input :value="name" @input="$emit(\'update:name\', $event.target.value)">',
        },
        'InstrumentReferenceFields': true,
        'InstrumentKeyboardFields': true,
        'InstrumentRegisterSection': true,
        'InstrumentBassSection': true,
        'UiFormSection': true,
        'v-tooltip': true,
      },
    },
  })
})

afterEach(() => {
  wrapper?.unmount()
  vi.clearAllTimers()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

async function edit(name: string) {
  await wrapper.get('input').setValue(name)
  await nextTick()
}

describe('accordion autosave drain', () => {
  it('coalesces overlapping edits and keeps navigation pending through the last durable write', async () => {
    const first = deferred()
    const second = deferred()
    backend.set.mockImplementationOnce(async (_id: string, instrument: Instrument) => {
      await first.promise
      saved = instrument
    }).mockImplementationOnce(async (_id: string, instrument: Instrument) => {
      await second.promise
      saved = instrument
    })
    await edit('First')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Second')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Latest')
    await vi.advanceTimersByTimeAsync(600)
    expect(backend.set).toHaveBeenCalledTimes(1)
    let navigated = false
    const navigation = flushEdits().then((ok) => {
      navigated = ok
    })
    first.resolve()
    await flushPromises()
    expect(backend.set).toHaveBeenCalledTimes(2)
    expect(backend.set.mock.calls[1]![1]).toMatchObject({ name: 'Latest', a4: 440.5 })
    expect(navigated).toBe(false)
    second.resolve()
    await navigation
    expect(navigated).toBe(true)
    expect(saved.name).toBe('Latest')
    expect(backend.save).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(600)
    expect(backend.set).toHaveBeenCalledTimes(2)
  })

  it('autosaves the newer edit without requiring navigation or another keystroke', async () => {
    const first = deferred()
    backend.set.mockReturnValueOnce(first.promise)
    await edit('First')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Latest')
    await vi.advanceTimersByTimeAsync(600)
    first.resolve()
    await flushPromises()
    expect(backend.set).toHaveBeenCalledTimes(2)
    expect(saved.name).toBe('Latest')
    expect(vi.getTimerCount()).toBe(0)
  })

  it('persists a revert made while an earlier snapshot is in flight', async () => {
    const first = deferred()
    backend.set.mockImplementationOnce(async (_id: string, instrument: Instrument) => {
      await first.promise
      saved = instrument
    })
    await edit('Changed')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Original')
    first.resolve()
    await flushPromises()
    expect(backend.set).toHaveBeenCalledTimes(2)
    expect(saved.name).toBe('Original')
  })

  it('drains a newer snapshot after rejection and clears the error on success', async () => {
    const first = deferred()
    backend.set.mockReturnValueOnce(first.promise)
    await edit('First')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Latest')
    await vi.advanceTimersByTimeAsync(600)
    first.reject(new Error('session.error.saveFailed'))
    await flushPromises()
    expect(backend.set).toHaveBeenCalledTimes(2)
    expect(saved.name).toBe('Latest')
    expect(error.value).toBe('')
  })

  it('keeps an unchanged failure visible and allows an explicit retry without a retry loop', async () => {
    backend.save.mockRejectedValueOnce(new Error('session.error.saveFailed'))
    await edit('Changed')
    await vi.advanceTimersByTimeAsync(600)
    await vi.advanceTimersByTimeAsync(3000)
    expect(error.value).toBe('session.error.saveFailed')
    expect(backend.set).toHaveBeenCalledTimes(1)
    expect(await flushEdits()).toBe(true)
    expect(backend.set).toHaveBeenCalledTimes(2)
    expect(error.value).toBe('')
  })

  it('stops draining when the panel unmounts during a write', async () => {
    const first = deferred()
    backend.set.mockReturnValueOnce(first.promise)
    await edit('First')
    await vi.advanceTimersByTimeAsync(600)
    await edit('Latest')
    wrapper.unmount()
    first.resolve()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(600)
    expect(backend.set).toHaveBeenCalledTimes(1)
  })
})
