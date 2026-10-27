// tests/sessionDialog.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { createVuetify } from 'vuetify'
import { VBtn, VTextField } from 'vuetify/components'
import SessionDialog from '~/components/session/SessionDialog.vue'
import AccordionQuickForm from '~/components/session/AccordionQuickForm.vue'
import FormField from '~/components/ui/FormField.vue'
import { useDraftNumber } from '~/composables/useDraftNumber'
import type { InstrumentTemplate } from '~/types/session'

const library = {
  list: ref<InstrumentTemplate[]>([]),
  refresh: vi.fn(),
  find: vi.fn(),
  save: vi.fn(),
}
let wrapper: ReturnType<typeof mount>

beforeEach(() => {
  library.list.value = []
  library.refresh.mockReset().mockResolvedValue(undefined)
  library.find.mockReset()
  library.save.mockReset().mockImplementation(async (item: InstrumentTemplate) => ({ ...item, id: 'new-instrument' }))
  vi.stubGlobal('useI18n', () => ({
    t: (key: string) => key,
    d: (date: Date) => date.toISOString().slice(0, 10),
  }))
  vi.stubGlobal('useInstruments', () => library)
  vi.stubGlobal('useDraftNumber', useDraftNumber)
})

afterEach(() => {
  wrapper?.unmount()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

async function open() {
  wrapper = mount(SessionDialog, {
    props: { modelValue: false },
    global: {
      plugins: [createVuetify({ components: { VBtn, VTextField } })],
      components: { UiFormField: FormField, SessionAccordionQuickForm: AccordionQuickForm },
      stubs: {
        'v-dialog': { template: '<div><slot /></div>' },
        'v-tooltip': true,
        'UiDialogCard': { template: '<div><slot /><slot name="actions" /></div>' },
        'SessionAccordionPicker': {
          emits: ['describe'],
          template: '<button class="describe" @click="$emit(\'describe\')">Describe</button>',
        },
      },
    },
  })
  await wrapper.setProps({ modelValue: true })
  await flushPromises()
}

const name = () => wrapper.get('input[aria-label="session.dialog.name"]')
const accordionName = () => wrapper.get('input[aria-label="instrument.name"]')
const start = () => wrapper.findAllComponents(VBtn).find(button => button.text() === 'session.dialog.start')!

describe('new session dialog', () => {
  it.each(['', '   '])('rejects blank instrument name %j in the field before saving', async (blank) => {
    await open()
    await wrapper.get('.describe').trigger('click')
    expect(wrapper.text()).not.toContain('instrument.nameRequired')
    await accordionName().setValue(blank)
    await start().trigger('click')
    await flushPromises()
    expect(library.save).not.toHaveBeenCalled()
    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('instrument.nameRequired')
    expect(accordionName().attributes('aria-invalid')).toBe('true')

    await accordionName().setValue('Morino')
    expect(wrapper.text()).not.toContain('instrument.nameRequired')
    await start().trigger('click')
    await flushPromises()
    expect(library.save).toHaveBeenCalledTimes(1)
    expect(wrapper.emitted('submit')).toHaveLength(1)
  })

  it.each([
    ['session.error.invalidInstrument', 'session.error.invalidInstrument'],
    ['unexpected internal detail', 'session.error.unexpected'],
  ])('shows a localized service failure and preserves the draft for retry: %s', async (failure, key) => {
    library.save.mockRejectedValueOnce(new Error(failure))
    await open()
    await wrapper.get('.describe').trigger('click')
    await accordionName().setValue('  Morino  ')
    await name().setValue('My session')
    await wrapper.get('input[aria-label="instrument.a4"]').setValue('440.5')
    await start().trigger('click')
    await flushPromises()
    expect(wrapper.get('.form__error[role="alert"]').text()).toBe(key)
    expect(wrapper.emitted('submit')).toBeUndefined()
    expect((accordionName().element as HTMLInputElement).value).toBe('  Morino  ')
    expect((name().element as HTMLInputElement).value).toBe('My session')
    expect(start().props('disabled')).toBe(false)

    await start().trigger('click')
    await flushPromises()
    expect(wrapper.find('.form__error').exists()).toBe(false)
    expect(library.save).toHaveBeenLastCalledWith(expect.objectContaining({
      name: 'Morino', instrument: expect.objectContaining({ a4: 440.5 }),
    }))
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      name: 'My session', instrumentId: 'new-instrument', instrument: { a4: 440.5 },
    })
  })

  it('does not duplicate an outstanding instrument save', async () => {
    let finish!: (item: InstrumentTemplate) => void
    library.save.mockImplementationOnce(() => new Promise<InstrumentTemplate>((resolve) => {
      finish = resolve
    }))
    await open()
    await wrapper.get('.describe').trigger('click')
    await accordionName().setValue('Morino')
    await start().trigger('click')
    await name().trigger('keyup', { key: 'Enter' })
    expect(library.save).toHaveBeenCalledTimes(1)
    finish({ ...library.save.mock.calls[0]![0], id: 'saved' })
    await flushPromises()
    expect(wrapper.emitted('submit')).toHaveLength(1)
  })

  it('resets a cancelled name when the offered name is unchanged', async () => {
    await open()
    const offered = (name().element as HTMLInputElement).value
    await name().setValue('Cancelled draft')
    await wrapper.setProps({ modelValue: false })
    await wrapper.setProps({ modelValue: true })
    await flushPromises()
    expect((name().element as HTMLInputElement).value).toBe(offered)
    await start().trigger('click')
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({ name: offered })
  })

  it('refreshes the suggested date and clears validation when reopened', async () => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2026-10-26T12:00:00Z'))
    await open()
    await wrapper.get('.describe').trigger('click')
    await start().trigger('click')
    expect(wrapper.text()).toContain('instrument.nameRequired')
    await wrapper.setProps({ modelValue: false })
    vi.setSystemTime(new Date('2026-10-27T12:00:00Z'))
    await wrapper.setProps({ modelValue: true })
    await flushPromises()
    expect((name().element as HTMLInputElement).value).toContain('2026-10-27')
    await wrapper.get('.describe').trigger('click')
    expect(wrapper.text()).not.toContain('instrument.nameRequired')
  })
})
