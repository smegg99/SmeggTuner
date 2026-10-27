import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import AccordionPicker from '~/components/session/AccordionPicker.vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import type { Bank, InstrumentTemplate } from '~/types/session'

/*
 * The one control that replaced a whole room. It has to do three things a plain
 * select does not: say what each accordion sounds, keep "describe a new one" and
 * "load from file" out of the model, and let one be forgotten without picking it.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any
g.useI18n = () => ({ t: (key: string) => key })

const global = {
  plugins: [createVuetify({ components })],
  stubs: {
    'v-menu': { template: '<div><slot :props="{}" name="activator" /><slot /></div>' },
    'v-icon': true,
  },
}

function template(id: string, name: string, banks: Bank[]): InstrumentTemplate {
  return {
    id,
    name,
    hasImage: false,
    instrument: { name, serial: '', reedCount: banks.length, banks },
  }
}

const ITEMS = [
  template('a', 'Bugari 508/ARS', ['L', 'M1', 'M2', 'H']),
  template('b', 'Hohner Gola', ['L', 'M1', 'M2']),
]

describe('the accordion picker', () => {
  it('lists every accordion already tuned, and what each one sounds', () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: ITEMS }, global })
    const text = w.text()

    expect(text).toContain('Bugari 508/ARS')
    expect(text).toContain('16\'+8\'+8\'+4\'')
    expect(text).toContain('Hohner Gola')
    expect(text).toContain('16\'+8\'+8\'')
  })

  it('picking one hands back its id', async () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: ITEMS }, global })
    await w.find('[data-test="pick-a"]').trigger('click')

    expect(w.emitted('update:modelValue')?.[0]).toEqual(['a'])
  })

  it('says which one is in hand', () => {
    const w = mount(AccordionPicker, { props: { modelValue: 'b', items: ITEMS }, global })
    expect(w.find('[data-test="current"]').text()).toBe('Hohner Gola')
  })

  it('says nothing is picked yet rather than showing a blank', () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: ITEMS }, global })
    expect(w.find('[data-test="current"]').text()).toBe('session.dialog.pickAccordion')
  })

  // Describing and loading are their own acts. If they rode in the model, the
  // dialog would carry a template id that names no accordion.
  it('keeps describe and load out of the model', async () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: ITEMS }, global })
    await w.find('[data-test="describe"]').trigger('click')
    await w.find('[data-test="load"]').trigger('click')

    expect(w.emitted('describe')).toHaveLength(1)
    expect(w.emitted('load')).toHaveLength(1)
    expect(w.emitted('update:modelValue')).toBeUndefined()
  })

  it('forgetting an accordion does not also pick it', async () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: ITEMS }, global })
    await w.find('[data-test="forget-b"]').trigger('click')

    expect(w.emitted('remove')?.[0]).toEqual(['b'])
    expect(w.emitted('update:modelValue')).toBeUndefined()
  })

  it('offers to describe one on a bench that has tuned nothing yet', () => {
    const w = mount(AccordionPicker, { props: { modelValue: '', items: [] }, global })

    expect(w.find('[data-test="describe"]').exists()).toBe(true)
    expect(w.findAll('[data-test^="pick-"]')).toHaveLength(0)
  })

  it('names an accordion nobody named, rather than drawing an empty row', () => {
    const w = mount(AccordionPicker, {
      props: { modelValue: '', items: [template('c', '', ['M1'])] },
      global,
    })
    expect(w.find('[data-test="pick-c"]').text()).toContain('session.card.unnamedInstrument')
  })
})
