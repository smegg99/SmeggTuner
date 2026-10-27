// tests/textInput.spec.ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { createVuetify } from 'vuetify'
import { VBtn, VTextField } from 'vuetify/components'
import SessionDialog from '~/components/session/SessionDialog.vue'
import { useDraftNumber } from '~/composables/useDraftNumber'

beforeEach(() => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key, d: () => '2026-10-24' }))
  vi.stubGlobal('useDraftNumber', useDraftNumber)
  vi.stubGlobal('useInstruments', () => ({ list: ref([]), find: () => undefined, refresh: vi.fn() }))
})

function dialog() {
  return mount(SessionDialog, {
    props: { modelValue: true },
    global: {
      plugins: [createVuetify({ components: { VBtn, VTextField } })],
      stubs: {
        VDialog: { template: '<div><slot /></div>' },
        UiDialogCard: { template: '<div><slot /><slot name="actions" /></div>' },
        UiFormField: { template: '<div><slot /></div>' },
        SessionAccordionPicker: true,
        SessionAccordionQuickForm: true,
      },
    },
  })
}

describe('new session name', () => {
  it('does not create a session when focus moves to another field', async () => {
    const w = dialog()
    await w.get('input').setValue('My accordion')
    await w.get('input').trigger('blur')
    expect(w.emitted('submit')).toBeUndefined()
  })

  it('submits the typed name on Enter', async () => {
    const w = dialog()
    await w.get('input').setValue('My accordion')
    await w.get('input').trigger('keyup', { key: 'Enter' })
    await flushPromises()
    expect(w.emitted('submit')?.[0]?.[0]).toMatchObject({ name: 'My accordion', instrument: { banks: [], reedCount: 1 } })
  })

  it('ignores Enter while the parent is saving the session', async () => {
    const w = dialog()
    await w.setProps({ saving: true })
    await w.get('input').trigger('keyup', { key: 'Enter' })
    expect(w.emitted('submit')).toBeUndefined()
    await w.setProps({ saving: false })
    await w.get('input').trigger('keyup', { key: 'Enter' })
    expect(w.emitted('submit')).toHaveLength(1)
  })

  it('remembers the instrument once when Enter is repeated during the save', async () => {
    let finish!: (value: undefined) => void
    const save = vi.fn(() => new Promise<undefined>((resolve) => {
      finish = resolve
    }))
    vi.stubGlobal('useInstruments', () => ({ list: ref([]), find: () => undefined, refresh: vi.fn(), save }))
    const w = dialog()
    w.findComponent({ name: 'SessionAccordionPicker' }).vm.$emit('describe')
    await flushPromises()
    w.findComponent({ name: 'SessionAccordionQuickForm' }).vm.$emit('update:name', 'Morino')
    await flushPromises()
    await w.get('input').trigger('keyup', { key: 'Enter' })
    await w.get('input').trigger('keyup', { key: 'Enter' })
    expect(save).toHaveBeenCalledTimes(1)
    finish(undefined)
    await flushPromises()
    expect(w.emitted('submit')).toBeUndefined()
  })
})
