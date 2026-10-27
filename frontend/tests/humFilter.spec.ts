import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import FilterToggles from '~/components/settings/audio/FilterToggles.vue'

/*
 * THE MAINS FILTER IS ONE CHOICE OF THREE.
 *
 * A supply is 50 Hz or 60 Hz, or the microphone is not hearing it. The config keeps a flag per
 * frequency, which is two booleans able to say four things - including "both", which is not a
 * supply. This component is the one place that knows they are alternatives.
 *
 * Off used to be reachable only by turning the lit one back off: there was no None to pick.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any
g.useI18n = () => ({ t: (key: string) => key })

const config = reactive({ audio: { hum_filter_50: false, hum_filter_60: false } })
const running = ref(false)

vi.mock('~/composables/useConfigSync', () => ({ useConfigSync: () => ({ config }) }))
vi.mock('~/composables/useTuner', () => ({ useTuner: () => ({ running }) }))

const global = {
  stubs: {
    UiFormField: { template: '<div><slot :id="\'x\'" /></div>' },
    UiToolGroup: {
      props: ['modelValue', 'items', 'disabled'],
      emits: ['update:modelValue'],
      template: '<span><button v-for="i in items" :key="i.value" :data-hz="i.value" :data-on="i.value === modelValue" :disabled="disabled" @click="$emit(\'update:modelValue\', i.value)">{{ i.label }}</button></span>',
    },
  },
}

function filters(fifty = false, sixty = false, live = false) {
  config.audio.hum_filter_50 = fifty
  config.audio.hum_filter_60 = sixty
  running.value = live
  return mount(FilterToggles, { global })
}

describe('the mains hum filter', () => {
  it('offers none as a choice of its own', () => {
    const w = filters()
    expect(w.findAll('button').map(b => b.attributes('data-hz'))).toEqual(['0', '50', '60'])
  })

  it('starts on none when neither notch is set', () => {
    expect(filters().find('[data-hz="0"]').attributes('data-on')).toBe('true')
  })

  it('picking a frequency sets that notch and only that one', async () => {
    const w = filters()
    await w.find('[data-hz="60"]').trigger('click')

    expect(config.audio.hum_filter_60).toBe(true)
    expect(config.audio.hum_filter_50).toBe(false)
  })

  it('picking none clears both', async () => {
    const w = filters(true)
    await w.find('[data-hz="0"]').trigger('click')

    expect(config.audio.hum_filter_50).toBe(false)
    expect(config.audio.hum_filter_60).toBe(false)
  })

  it('switching frequency does not leave the old notch on', async () => {
    const w = filters(true)
    await w.find('[data-hz="60"]').trigger('click')

    expect(config.audio.hum_filter_50).toBe(false)
    expect(config.audio.hum_filter_60).toBe(true)
  })

  // A hand-edited config can say both. Two notches are not a supply; it reads as the lower one
  // and the next change writes a state that means something.
  it('reads a config claiming both as 50', async () => {
    const w = filters(true, true)
    expect(w.find('[data-hz="50"]').attributes('data-on')).toBe('true')

    await w.find('[data-hz="50"]').trigger('click')
    expect(config.audio.hum_filter_60).toBe(false)
  })

  // The notches are applied at engine start, so they cannot be moved under a running engine.
  it('is locked while the engine runs', () => {
    const w = filters(false, false, true)
    expect(w.findAll('button').every(b => b.attributes('disabled') !== undefined)).toBe(true)
  })
})
