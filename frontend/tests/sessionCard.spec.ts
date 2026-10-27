// tests/sessionCard.spec.ts
import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import ToolKey from '@s99ab/components/primitives/ToolKey.vue'
import SessionCard from '~/components/session/SessionCard.vue'
import type { Summary } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'

// The shelf card: its face opens the session, its own keys export and delete it.
vi.mock('~/composables/useInstruments', () => ({ useInstruments: () => ({ find: () => undefined }) }))
const t = (key: string) => key
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t, d: () => '' })

function card(src = '') {
  const session = { id: 's1', name: 'Paolo Soprani', instrument: { name: 'Studio 3', reedCount: 3, banks: [] }, readings: 4 } as unknown as Summary
  return mount(SessionCard, {
    props: { session, src },
    global: {
      plugins: [createVuetify({ components })],
      components: { UiToolKey: ToolKey },
      stubs: { SessionCardMeta: true, VIcon: { template: '<i class="icon" />' } },
    },
  })
}

describe('a session card', () => {
  it('shows the session and its accordion, and opens on its face', async () => {
    const w = card()
    expect(w.find('.card__name').text()).toBe('Paolo Soprani')
    expect(w.find('.card__what').text()).toBe('Studio 3')
    expect(w.find('.card__img').exists()).toBe(false)
    await w.find('.card__open').trigger('click')
    expect(w.emitted('open')).toEqual([['s1']])
  })

  it('shows a photograph when there is one', () => {
    expect(card('/instruments/i1/image').find('.card__img').attributes('src')).toBe('/instruments/i1/image')
  })

  it('exports and deletes from its own keys, not by opening', async () => {
    const w = card()
    const [exp, del] = w.findAll('.card__keys button')
    await exp!.trigger('click')
    await del!.trigger('click')
    expect(w.emitted('export')).toEqual([['s1']])
    expect(w.emitted('delete')).toEqual([['s1']])
    expect(w.emitted('open')).toBeUndefined()
  })
})
