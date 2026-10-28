// tests/fileShortcuts.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import FileTransport from '~/components/file/FileTransport.vue'
import { useShortcuts } from '~/composables/useShortcuts'
import { fit, fitSelection, state, view, zoom } from '~/composables/transportState'

const host = vi.hoisted(() => ({
  room: 'calibrate', kind: 'file', running: true, starting: false, muted: false,
  start: vi.fn(), pause: vi.fn(), seek: vi.fn(), loop: vi.fn(), mute: vi.fn(),
}))
vi.mock('~/composables/useShell', () => ({
  useShell: () => ({ view: { get value() { return host.room } }, setView: vi.fn() }),
}))
vi.mock('~/composables/useAudioDevices', () => ({
  useAudioDevices: () => ({ current: { get value() { return { kind: host.kind } } } }),
}))
vi.mock('~/composables/useTuner', () => ({
  AUTO_NOTE: -1,
  useTuner: () => ({
    running: { get value() { return host.running } },
    starting: { get value() { return host.starting } },
    start: host.start, manualNote: ref(-1), setAuto: vi.fn(), stepNote: vi.fn(),
    hold: ref(false), setHold: vi.fn(), takeNow: vi.fn(),
  }),
}))
vi.mock('~/composables/useRecord', () => ({
  useRecord: () => ({ sessionId: ref(null), armed: ref(false), readings: ref(0), recording: ref(false) }),
}))
vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({
    state, view, zoom, fit, fitSelection, hasSelection: ref(false), clearSelection: vi.fn(),
    seek: host.seek, setPaused: host.pause, setLoop: host.loop,
  }),
}))
vi.mock('~/composables/useOutputLevel', () => ({
  useOutputLevel: () => ({
    volume: ref(1), muted: { get value() { return host.muted } }, setVolume: vi.fn(), setMuted: host.mute,
  }),
}))

const shell = defineComponent({
  setup() {
    useShortcuts()
    return () => h('div', [
      h('div', { style: 'display: none' }, [h(FileTransport)]),
      h(FileTransport, { withVolume: true }),
      h('button', { id: 'control' }), h('input'), h('select'),
      h('div', { contenteditable: 'true' }, [h('span')]),
      h('div', { role: 'slider', tabindex: 0 }),
    ])
  },
})

describe('one shell owner for file shortcuts', () => {
  let wrapper: ReturnType<typeof mount>
  beforeEach(() => {
    vi.clearAllMocks()
    Object.assign(host, { room: 'calibrate', kind: 'file', running: true, starting: false, muted: false })
    Object.assign(state, { available: true, duration: 100, from: 10, to: 90, paused: false, loop: false })
    Object.assign(view, { from: 0, to: 100 })
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
    wrapper = mount(shell, {
      attachTo: document.body,
      global: { stubs: { VBtn: true, UiToolKey: true, UiLevelControl: true } },
    })
  })
  afterEach(() => {
    wrapper.unmount()
    vi.unstubAllGlobals()
    document.body.replaceChildren()
  })

  function press(key: string, target: Element = wrapper.element, options: KeyboardEventInit = {}) {
    const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...options })
    target.dispatchEvent(event)
    return event
  }

  it('zooms only once with hidden tuner and visible calibration transports mounted', () => {
    expect(press('+').defaultPrevented).toBe(true)
    expect(view.to - view.from).toBe(50)
    expect([view.from, view.to]).toEqual([25, 75])
  })

  it.each(['settings', 'workshop'])('ignores hidden transport in %s', (room) => {
    host.room = room
    for (const key of [' ', '+', 'Home', 'End', 'l', 'm', 'f', 's']) {
      expect(press(key).defaultPrevented).toBe(false)
    }
    expect(view.to - view.from).toBe(100)
    expect(host.pause).not.toHaveBeenCalled()
    expect(host.seek).not.toHaveBeenCalled()
    expect(host.loop).not.toHaveBeenCalled()
    expect(host.mute).not.toHaveBeenCalled()
  })

  it('requires selected file input and an available transport', () => {
    host.kind = 'microphone'
    expect(press('+').defaultPrevented).toBe(false)
    host.kind = 'file'
    state.available = false
    expect(press('+').defaultPrevented).toBe(false)
    expect(view.to - view.from).toBe(100)
  })

  it('keeps keys with controls, modifiers, composition, dialogs and prior handlers', () => {
    for (const options of [{ ctrlKey: true }, { altKey: true }, { metaKey: true }, { isComposing: true }]) {
      expect(press('+', wrapper.element, options).defaultPrevented).toBe(false)
    }
    for (const selector of ['input', 'select', '[contenteditable] span']) {
      expect(press('+', wrapper.get(selector).element).defaultPrevented).toBe(false)
    }
    expect(press(' ', wrapper.get('#control').element).defaultPrevented).toBe(false)
    for (const key of ['Home', 'End']) {
      expect(press(key, wrapper.get('[role="slider"]').element).defaultPrevented).toBe(false)
    }
    const handled = new KeyboardEvent('keydown', { key: '+', bubbles: true, cancelable: true })
    handled.preventDefault()
    wrapper.element.dispatchEvent(handled)
    const overlay = document.createElement('div')
    overlay.className = 'v-overlay--active v-dialog'
    document.body.append(overlay)
    expect(press('+').defaultPrevented).toBe(false)
    expect(view.to - view.from).toBe(100)
    expect(host.pause).not.toHaveBeenCalled()
    overlay.className = 'v-overlay--active v-tooltip'
    press('+')
    expect(view.to - view.from).toBe(50)
  })

  it('starts a stopped file after clearing pause and ignores repeats or an active start', async () => {
    host.room = 'tune'
    host.running = false
    state.paused = true
    press(' ')
    await flushPromises()
    expect(host.pause).toHaveBeenCalledExactlyOnceWith(false)
    expect(host.start).toHaveBeenCalledOnce()
    expect(host.pause.mock.invocationCallOrder[0]).toBeLessThan(host.start.mock.invocationCallOrder[0]!)
    press(' ', wrapper.element, { repeat: true })
    host.starting = true
    press(' ')
    await flushPromises()
    expect(host.start).toHaveBeenCalledOnce()
  })

  it('runs each transport action once and permits repeated zoom only', () => {
    press(' ')
    press('Home')
    press('End')
    press('L')
    press('m')
    for (const key of [' ', 'l', 'm']) press(key, wrapper.element, { repeat: true })
    expect(host.pause).toHaveBeenCalledExactlyOnceWith(true)
    expect(host.seek.mock.calls).toEqual([[10], [90]])
    expect(host.loop).toHaveBeenCalledExactlyOnceWith(true)
    expect(host.mute).toHaveBeenCalledExactlyOnceWith(true)
    press('s')
    expect([view.from, view.to]).toEqual([10, 90])
    press('f')
    expect([view.from, view.to]).toEqual([0, 100])
    press('=', wrapper.element, { repeat: true })
    expect(view.to - view.from).toBe(50)
    press('-')
    expect(view.to - view.from).toBe(100)
  })

  it('removes the only shortcut listener when the shell unmounts', () => {
    wrapper.unmount()
    expect(press('+', document.body).defaultPrevented).toBe(false)
    expect(view.to - view.from).toBe(100)
  })
})
