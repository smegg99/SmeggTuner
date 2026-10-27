import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ValueBox from '~/components/tuner/ValueBox.vue'
import DeviationBar from '~/components/tuner/DeviationBar.vue'
import { buildBoxes } from '~/utils/boxes'
import type { Box, BoxInput } from '~/utils/boxes'

/*
 * useI18n is auto-imported by the Nuxt module and there is no Nuxt here, so it is
 * stubbed to return the key. That is enough: what is under test is not the wording,
 * it is that NO DIGIT REACHES THE DOM for a slot the engine did not hear, and that the
 * unit a number is labelled with is the unit it is in.
 */
const global = {
  components: { TunerDeviationBar: DeviationBar },
  mocks: { t: (key: string) => key },
  stubs: { TunerDeviationBar: DeviationBar },
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t: (key: string) => key })

const mountBox = (box: Box) => mount(ValueBox, { props: { box }, global })

const input: BoxInput = {
  panels: [],
  reedErrors: [],
  beatErrors: [],
  tolerance: 3,
  beatTolerance: 3,
  reedCount: 2,
  unit: 'reference',
  beatsInHz: true,
  relative: false,
}

const reed: Box = {
  kind: 'reed',
  index: 0,
  unit: 'reference',
  value: 4.2,
  goal: -15.8,
  decimals: 1,
  signed: true,
  dimmed: false,
  blank: null,
  out: true,
  frac: 0.675,
}

describe('the value box', () => {
  /*
   * A slot the engine measured and did not hear has no number. The reference keeps the sentinel
   * -99 behind it; it must reach the DOM by no route: not as a digit, not as a goal, not as a bar
   * a position could be read off.
   */
  it('renders no digits at all for a slot the engine did not hear', () => {
    const boxes = buildBoxes({
      ...input,
      panels: [
        { value: -99, target: -15.8, signed: true, decimals: 1, dimmed: false, shown: '-' },
        { value: -99, target: -99, signed: false, decimals: 1, dimmed: false, shown: '-' },
        { value: 1.4, target: 15.9, signed: true, decimals: 1, dimmed: false, shown: '+1.4' },
      ],
    })

    const wrapper = mountBox(boxes[0]!)

    // the box says WHY it is empty, in words
    expect(wrapper.find('.box__value--none').exists()).toBe(true)
    expect(wrapper.find('.box__value').text()).toBe('tuner.boxes.notHeard')

    // The scale is still drawn, but there is NO BAR ON IT: a bar is a position, and a position is a reading.
    expect(wrapper.findComponent(DeviationBar).exists()).toBe(true)
    expect(wrapper.findComponent(DeviationBar).props('frac')).toBeNull()
    expect(wrapper.find('.dev__bar').exists()).toBe(false)

    expect(wrapper.text()).not.toContain('99')
    expect(wrapper.text()).not.toContain('15.8')
  })

  it('prints a beat in hertz at the readout\'s own precision, labelled as hertz', () => {
    const boxes = buildBoxes({
      ...input,
      panels: [
        { value: 1, target: -99, signed: true, decimals: 1, dimmed: false, shown: '+1.0' },
        { value: 2.34, target: -99, signed: false, decimals: 2, dimmed: false, shown: '2.34' },
        { value: 1, target: -99, signed: true, decimals: 1, dimmed: false, shown: '+1.0' },
      ],
    })

    const wrapper = mountBox(boxes[1]!)
    expect(wrapper.find('.box__value').text()).toBe('2.34')
    expect(wrapper.find('.box__unit').text()).toBe('tuner.boxes.unit.hz')
  })

  it('prints a reed the engine did measure, and marks the backend verdict', () => {
    const wrapper = mountBox(reed)

    expect(wrapper.find('.box__value').text()).toBe('+4.2')
    expect(wrapper.find('.box__unit').text()).toBe('tuner.boxes.unit.reference')
    expect(wrapper.classes()).toContain('box--out')
    expect(wrapper.findComponent(DeviationBar).props('out')).toBe(true)
  })

  it('does not colour a reed the backend calls in tolerance, however large it looks', () => {
    const wrapper = mountBox({ ...reed, value: 9, goal: 0, out: false, frac: 0.875 })
    expect(wrapper.classes()).not.toContain('box--out')
  })

  // No note: the number stays, dimmed, as the reference keeps it.
  it('keeps a dimmed number on screen, dimmed', () => {
    const wrapper = mountBox({ ...reed, out: false, dimmed: true })
    expect(wrapper.find('.box__value').text()).toBe('+4.2')
    expect(wrapper.classes()).toContain('box--dim')
    expect(wrapper.classes()).not.toContain('box--out')
  })
})
