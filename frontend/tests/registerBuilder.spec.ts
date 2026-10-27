import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import RegisterBuilder from '~/components/instrument/RegisterBuilder.vue'
import type { Register } from '~/types/session'

/*
 * BUILDING A SWITCH.
 *
 * There are no common-switch presets any more: every switch on the card is built here, so this
 * is the only way ranks ever reach an instrument.
 *
 * The 8' count fills IN ORDER - 3 means M1..M3. Which rank is which is the card's column order,
 * and a gap would file a reading under a column nobody meant.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any
g.useI18n = () => ({ t: (key: string) => key })

const global = {
  stubs: {
    'v-icon': true,
    'InstrumentRegisterSymbol': true,
    // No $emit('click'): the real UiToolKey has no defineEmits and the parent's handler reaches
    // its root button by attribute fallthrough. A stub that emits as well fires every click twice.
    'UiToolKey': {
      props: ['label', 'active', 'disabled', 'icon'],
      template: '<button :data-active="active" :disabled="disabled">{{ label }}</button>',
    },
    // The 8' count is a segmented group; the test drives it through its model.
    'UiToolGroup': {
      props: ['modelValue', 'items'],
      emits: ['update:modelValue'],
      template: '<span class="counts"><button v-for="i in items" :key="i.value" :data-count="i.value" :data-on="i.value === modelValue" @click="$emit(\'update:modelValue\', i.value)">{{ i.label }}</button></span>',
    },
  },
}

function builder(have: Register[] = []) {
  return mount(RegisterBuilder, { props: { have }, global })
}

function added(w: ReturnType<typeof builder>) {
  return w.emitted('add')?.map(e => (e as [Register])[0])
}

// The rows mirror the engraving: 4' on top, the 8' count in the middle, 16' at the bottom.
const high = (w: ReturnType<typeof builder>) => w.findAll('.build__row button')[0]!
const low = (w: ReturnType<typeof builder>) => w.findAll('.build__row button').at(-1)!
const count = (w: ReturnType<typeof builder>, n: number) => w.find(`[data-count="${n}"]`)
const addKey = (w: ReturnType<typeof builder>) => w.findAll('.build__body > button').at(-1)!

describe('building a switch', () => {
  it('cannot be added before a rank is picked', () => {
    expect(addKey(builder()).attributes('disabled')).toBeDefined()
  })

  it('fills the 8 foot ranks in order', async () => {
    const w = builder()
    await count(w, 3).trigger('click')
    await addKey(w).trigger('click')

    expect(added(w)![0]!.banks).toEqual(['M1', 'M2', 'M3'])
  })

  it('names what it built the way the card names its columns', async () => {
    const w = builder()
    await low(w).trigger('click')
    await count(w, 2).trigger('click')
    await high(w).trigger('click')
    await addKey(w).trigger('click')

    expect(added(w)![0]).toEqual({ name: 'LMMH', banks: ['L', 'M1', 'M2', 'H'] })
  })

  it('shows the count in force', async () => {
    const w = builder()
    await count(w, 2).trigger('click')

    expect(count(w, 2).attributes('data-on')).toBe('true')
    expect(count(w, 3).attributes('data-on')).toBe('false')
  })

  it('clears itself once the switch is added', async () => {
    const w = builder()
    await low(w).trigger('click')
    await addKey(w).trigger('click')

    expect(low(w).attributes('data-active')).toBe('false')
    expect(addKey(w).attributes('disabled')).toBeDefined()
  })

  // The same rule core/session enforces: a switch IS the ranks it sounds.
  it('cannot add one the accordion already has', async () => {
    const w = builder([{ name: 'L', banks: ['L'] }])
    await low(w).trigger('click')

    expect(addKey(w).attributes('disabled')).toBeDefined()
    await addKey(w).trigger('click')
    expect(added(w)).toBeUndefined()
  })
})
