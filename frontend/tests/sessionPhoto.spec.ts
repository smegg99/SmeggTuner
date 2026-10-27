import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { computed, ref } from 'vue'
import SessionIdent from '~/components/session/SessionIdent.vue'

/*
 * THE PHOTOGRAPH IS A PICTURE UNTIL IT IS A PICKER.
 *
 * The session header carries the accordion's photograph on every tab. Setting one is an
 * accordion edit, so it belongs to the Accordion tab and nowhere else - and a photograph only
 * has somewhere to live once the accordion is on the shelf.
 *
 * It used to be a <button> on all three tabs, disabled outside the Accordion one but still
 * announcing "Photograph" on hover. A control that names itself and then does nothing reads as
 * broken, so outside its own tab the tile is a plain span: no button, no tooltip, no hover.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any
g.useI18n = () => ({ t: (key: string) => key })

const setPhoto = vi.fn()

// Whether the open session's accordion is on the shelf; the tests below drive it.
const remembered = ref(true)

vi.mock('~/composables/useSessionAccordion', () => ({
  useSessionAccordion: () => ({
    source: computed(() => ({ id: remembered.value ? 'a' : '', name: 'Bugari', hasImage: false })),
    remembered: computed(() => remembered.value),
    imageSrc: computed(() => ''),
    setPhoto,
    exportIt: vi.fn(),
    remember: vi.fn(),
  }),
}))

const global = {
  stubs: {
    'v-icon': true,
    'v-tooltip': true,
    'UiToolGroup': true,
    'UiToolKey': true,
    'UiChip': true,
  },
}

const SESSION = { id: 's1', name: 'Bugari - 19 Aug 2026', a4: 442, instrument: { name: 'Bugari' } }

function ident(view: 'recorded' | 'curve' | 'accordion', onShelf = true) {
  remembered.value = onShelf
  setPhoto.mockClear()
  return mount(SessionIdent, {
    props: { session: SESSION as never, busy: false, view },
    global,
  })
}

describe('the photograph in the session header', () => {
  for (const view of ['recorded', 'curve'] as const) {
    it(`is a plain picture on the ${view} tab, not a control`, async () => {
      const w = ident(view)
      const tile = w.find('.ident__pic')

      expect(tile.element.tagName).toBe('SPAN')
      expect(tile.attributes('title')).toBeUndefined()
      expect(tile.classes()).not.toContain('ident__pic--live')

      await tile.trigger('click')
      expect(setPhoto).not.toHaveBeenCalled()
    })
  }

  it('is the picker on the accordion tab', async () => {
    const w = ident('accordion')
    const tile = w.find('.ident__pic')

    expect(tile.element.tagName).toBe('BUTTON')
    expect(tile.attributes('title')).toBe('instrument.photo')
    expect(tile.classes()).toContain('ident__pic--live')

    await tile.trigger('click')
    expect(setPhoto).toHaveBeenCalledTimes(1)
  })

  // A photograph is stored on the shelf entry. Until the accordion has one there is nowhere to
  // put the image, so the tile stays a picture even on its own tab.
  it('stays a picture while the accordion is not on the shelf', async () => {
    const w = ident('accordion', false)
    const tile = w.find('.ident__pic')

    expect(tile.element.tagName).toBe('SPAN')
    await tile.trigger('click')
    expect(setPhoto).not.toHaveBeenCalled()
  })
})
