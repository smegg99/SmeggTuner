import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SessionLabel from '~/components/shell/SessionLabel.vue'
import type { SessionProgress } from '~/utils/sessionProgress'

/*
 * The footer's session label. useI18n is auto-imported by the Nuxt module and
 * there is no Nuxt here, so it is stubbed to return the key: what is under test is
 * that the NAME reaches the bar, not the wording around it.
 *
 * THE REGRESSION: an open session that is not being recorded is the ordinary
 * state - you open a box long before you arm a pass - and the label went blank
 * there, because the name was gated behind the recording ('active') branch. The
 * bar showed nothing for a session that was plainly open.
 */
const withStub = globalThis as typeof globalThis & { useI18n: () => { t: (key: string) => string } }
withStub.useI18n = () => ({ t: (key: string) => key })

const mountLabel = (progress: SessionProgress) => mount(SessionLabel, { props: { progress } })

describe('the footer session label', () => {
  it('shows the name of an open session that is not being recorded', () => {
    const w = mountLabel({ kind: 'idle', name: 'My Rossetti' })
    expect(w.find('.label__name').text()).toBe('My Rossetti')
    expect(w.findAll('.label__prog')).toHaveLength(0)
  })

  it('shows the name and both figures while a pass is recorded', () => {
    const w = mountLabel({ kind: 'active', name: 'My Rossetti', done: 23, total: 41, inTune: 18 })
    expect(w.find('.label__name').text()).toBe('My Rossetti')
    expect(w.findAll('.label__prog')).toHaveLength(2)
  })

  it('drops the in-tune figure but keeps name and coverage with no goal curve', () => {
    const w = mountLabel({ kind: 'active', name: 'My Rossetti', done: 23, total: 41, inTune: null })
    expect(w.find('.label__name').text()).toBe('My Rossetti')
    expect(w.findAll('.label__prog')).toHaveLength(1)
  })

  it('says "no session" and shows no name when nothing is open', () => {
    const w = mountLabel({ kind: 'none' })
    expect(w.find('.label__none').exists()).toBe(true)
    expect(w.find('.label__name').exists()).toBe(false)
  })
})
