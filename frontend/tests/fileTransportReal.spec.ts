import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import FileTransport from '~/components/file/FileTransport.vue'
import ToolKey from '@s99ab/components/primitives/ToolKey.vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import LevelControl from '~/components/ui/LevelControl.vue'

/*
 * The transport, mounted with its REAL children.
 *
 * The other spec stubs UiToolKey, UiToolGroup and UiLevelSlider, which is exactly why it
 * passed while the app showed an empty toolbar: a child that throws during setup takes
 * the whole subtree down with it and renders nothing, and a stub cannot throw. So this
 * one uses the real components and real Vuetify, and it fails if any of them cannot be
 * constructed.
 */

vi.mock('~/composables/useTuner', () => ({
  useTuner: () => ({
    running: ref(true),
    starting: ref(false),
    start: vi.fn(),
    stop: vi.fn(),
  }),
}))

vi.mock('~/composables/useOutputLevel', () => ({
  useOutputLevel: () => ({ volume: ref(1), muted: ref(false), setVolume: vi.fn(), setMuted: vi.fn() }),
}))

vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({
    state: reactive({
      available: true,
      name: 'h-16.wav',
      duration: 9,
      position: 2,
      from: 0,
      to: 9,
      paused: false,
      moving: true,
      loop: false,
      sampleRate: 48000,
    }),
    view: reactive({ from: 0, to: 9 }),
    hasSelection: ref(false),
    muted: ref(false),
    volume: ref(1),
    seek: vi.fn(),
    setPaused: vi.fn(),
    stop: vi.fn(),
    setLoop: vi.fn(),
    setMuted: vi.fn(),
    setVolume: vi.fn(),
    clearSelection: vi.fn(),
    zoom: vi.fn(),
    fit: vi.fn(),
    fitSelection: vi.fn(),
  }),
}))

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t: (key: string) => key })

/*
 * v-icon is stubbed and Vuetify is not installed: its component CSS cannot be loaded by
 * vitest, and the icon is not what is under test. The toolbar at the top of the app is
 * built from these same keys and its icons appear, so v-icon works. What has never been
 * constructed anywhere is ToolKey's and LevelSlider's own setup, and that is what this
 * mounts for real.
 */
/*
 * Every component the transport reaches is registered by name, including the ones it only
 * reaches THROUGH another: the app resolves these by Nuxt auto-import, and a test that
 * mounts by hand does not get that. UiLevelControl wraps UiToolKey and UiLevelSlider.
 *
 * A missing registration here does NOT show up as a mount that throws. Vue's compiler
 * hoists resolveComponent to the top of the render function, so the warning fires even for
 * a component behind a v-if that never runs - and the mount succeeds either way. Read the
 * assertions, not the warnings.
 */
const mountReal = (props: { withVolume?: boolean } = {}) =>
  mount(FileTransport, {
    props,
    global: {
      plugins: [createVuetify({ components })],
      components: {
        UiToolKey: ToolKey,
        UiLevelControl: LevelControl,
      },
      stubs: { VIcon: { template: '<i />' } },
      mocks: { t: (key: string) => key },
    },
  })

describe('the file transport, with its real children', () => {
  it('constructs every child without throwing', () => {
    expect(() => mountReal()).not.toThrow()
  })

  // The exact set matters. It renders NOTHING at all when the keys are handed to a
  // component that has no slot, which is what shipped, and a count is the only assertion
  // that catches that.
  it('puts every key on the screen', () => {
    const titles = mountReal().findAll('button').map(b => b.attributes('title'))

    expect(titles).toEqual([
      'file.toStart',
      'file.toEnd',
      'file.loop',
      'file.zoomOut',
      'file.zoomIn',
      'file.fit',
      'file.clearSelection',
    ])
  })

  /*
   * The volume is OPT-IN, and the default is no volume: the tune room's toolbar already
   * carries one output level and a second one under it would be two controls for one
   * thing. Only a room whose toolbar has no volume key of its own asks for it - see
   * CalibrationView, which is the sole caller passing with-volume.
   */
  it('carries no volume of its own by default', () => {
    const w = mountReal()

    expect(w.findAll('button').map(b => b.attributes('title'))).not.toContain('file.mute')
    expect(w.find('[role="slider"]').exists()).toBe(false)
  })

  it('puts the mute key in the row when the volume is asked for', () => {
    const titles = mountReal({ withVolume: true }).findAll('button').map(b => b.attributes('title'))

    expect(titles).toEqual([
      'file.toStart',
      'file.toEnd',
      'file.loop',
      'file.mute',
      'file.zoomOut',
      'file.zoomIn',
      'file.fit',
      'file.clearSelection',
    ])
  })

  /*
   * Play, pause and stop belong to FileRun, at the head of the same row. They are not
   * duplicated here, and they are not in the app toolbar either: starting the engine and
   * playing the recording are one action.
   */
  it('does not duplicate the run controls', () => {
    const titles = mountReal().findAll('button').map(b => b.attributes('title'))

    expect(titles).not.toContain('file.play')
    expect(titles).not.toContain('file.pause')
    expect(titles).not.toContain('file.stop')
  })

  it('puts a real volume control on the screen', () => {
    expect(mountReal({ withVolume: true }).find('[role="slider"]').exists()).toBe(true)
  })
})
