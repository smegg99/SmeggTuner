// tests/bassBuilder.spec.ts
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import ToolKey from '@s99ab/components/primitives/ToolKey.vue'
import BassBuilder from '~/components/instrument/BassBuilder.vue'
import FormField from '~/components/ui/FormField.vue'
import { MAX_BASS_REGISTERS } from '~/types/session'
import type { BassRegister } from '~/types/session'

// The bass switch builder: feet picked, a name typed, and Add appends the switch.
const t = (key: string) => key
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t })

function builder(registers: BassRegister[] = [], feet = [16, 8, 4]) {
  return mount(BassBuilder, {
    props: { 'modelValue': registers, feet, 'onUpdate:modelValue': () => {} },
    global: { plugins: [createVuetify({ components })], components: { UiToolKey: ToolKey, UiFormField: FormField } },
  })
}

const feetKeys = (w: ReturnType<typeof builder>) => w.findAll('.bass__feet button')
const addKey = (w: ReturnType<typeof builder>) => w.findAll('.bass__add button').at(-1)!

describe('the bass switch builder', () => {
  it('builds a switch from the feet picked and the name typed', async () => {
    const w = builder()
    expect(feetKeys(w).map(k => k.text())).toEqual(['16\'', '8\'', '4\''])
    await feetKeys(w)[2]!.trigger('click')
    await feetKeys(w)[0]!.trigger('click')
    await w.find('.bass__add input').setValue('Master')
    await addKey(w).trigger('click')
    expect(w.emitted('update:modelValue')).toEqual([[[{ name: 'Master', feet: [16, 4] }]]])
    expect((w.find('.bass__add input').element as HTMLInputElement).value).toBe('')
  })

  it('goes quiet when the machine is full', () => {
    const full = Array.from({ length: MAX_BASS_REGISTERS }, (_, i) => ({ name: `S${i}`, feet: [16] }))
    const w = builder(full)
    expect(feetKeys(w).every(k => k.attributes('disabled') !== undefined)).toBe(true)
  })

  it('drops the feet a smaller machine no longer has from the switch being built', async () => {
    const w = builder()
    await feetKeys(w)[0]!.trigger('click')
    await feetKeys(w)[2]!.trigger('click')
    await w.setProps({ feet: [8, 4] })
    await nextTick()
    await w.find('.bass__add input').setValue('Bassoon')
    await addKey(w).trigger('click')
    expect(w.emitted('update:modelValue')).toEqual([[[{ name: 'Bassoon', feet: [4] }]]])
  })
})
