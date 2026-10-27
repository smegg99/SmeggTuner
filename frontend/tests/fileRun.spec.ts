import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import FileRun from '~/components/file/FileRun.vue'
import { useFileRun } from '~/composables/useFileRun'
import ToolKey from '@s99ab/components/primitives/ToolKey.vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'

/*
 * Play, Pause, Stop - the transport as it was agreed.
 *
 * Starting the engine and playing the recording are ONE action, so Play IS Start. That is
 * the whole point of this component existing, and it is the thing that was wrong twice.
 */
const engine = { running: ref(false), starting: ref(false), start: vi.fn(), stop: vi.fn() }
const file = reactive({ paused: false, moving: true, from: 1.5, to: 4, available: true })
const setPaused = vi.fn()
const seek = vi.fn()

vi.mock('~/composables/useTuner', () => ({ useTuner: () => engine }))
vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({ state: file, setPaused, seek }),
}))

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t: (key: string) => key })

const mountRun = () =>
  mount(FileRun, {
    global: {
      plugins: [createVuetify({ components })],
      components: { UiToolKey: ToolKey },
      stubs: { VIcon: { template: '<i />' } },
      mocks: { t: (key: string) => key },
    },
  })

const keys = () => mountRun().findAll('button')

// The play key's faces and actions (useFileRun); the file card's first key is that play key.
describe('the recording run', () => {
  it('shows play when nothing is playing, pause once it plays, play again when paused', () => {
    const { playing } = useFileRun()
    engine.running.value = false
    file.paused = false
    expect(playing.value).toBe(false)
    engine.running.value = true
    expect(playing.value).toBe(true)
    file.paused = true
    expect(playing.value).toBe(false)
  })

  // PLAY IS START. There is no separate Start anywhere for a recording.
  it('starts the engine when nothing is running', async () => {
    engine.running.value = false
    engine.start.mockClear()
    await useFileRun().playPause()
    expect(engine.start).toHaveBeenCalled()
  })

  it('pauses instead of starting when the engine is already up', async () => {
    engine.running.value = true
    file.paused = false
    engine.start.mockClear()
    setPaused.mockClear()
    await useFileRun().playPause()
    expect(engine.start).not.toHaveBeenCalled()
    expect(setPaused).toHaveBeenCalledWith(true)
  })

  // A pause left over from the last run would make Play look broken: the engine would come up and play nothing.
  it('lifts a leftover pause when it starts', async () => {
    engine.running.value = false
    file.paused = true
    setPaused.mockClear()
    await useFileRun().playPause()
    expect(setPaused).toHaveBeenCalledWith(false)
  })
})

describe('the recording run keys', () => {
  it('offers exactly play and stop', () => {
    expect(keys()).toHaveLength(2)
  })

  it('shows pause on the play key while it plays', () => {
    engine.running.value = true
    file.paused = false
    expect(keys()[0]!.attributes('data-icon')).toBe('mdi-pause')
  })

  // STOP ends the run AND returns the needle to the top of the selection - not the top of
  // the file. The selection is the fragment being worked on.
  it('stops the engine and returns to the start of the selection', async () => {
    engine.running.value = true
    engine.stop.mockClear()
    seek.mockClear()

    await keys()[1]!.trigger('click')

    expect(engine.stop).toHaveBeenCalled()
    expect(seek).toHaveBeenCalledWith(1.5)
  })

  it('cannot be stopped when it is not running', () => {
    engine.running.value = false
    expect(keys()[1]!.attributes('disabled')).toBeDefined()
  })
})
