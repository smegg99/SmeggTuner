// tests/shortcutEvents.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import { useShortcuts } from '~/composables/useShortcuts'

const calls = vi.hoisted(() => ({
  undo: vi.fn(), take: vi.fn(), record: vi.fn(), step: vi.fn(), freeze: vi.fn(),
}))
vi.mock('~/composables/useAudioDevices', () => ({
  useAudioDevices: () => ({ current: ref({ kind: 'microphone' }) }),
}))
vi.mock('~/composables/useTransport', () => ({
  useTransport: () => ({ state: { available: false } }),
}))
vi.mock('~/composables/useOutputLevel', () => ({
  useOutputLevel: () => ({ muted: ref(false), setMuted: vi.fn() }),
}))
vi.mock('~/composables/useShell', () => ({
  useShell: () => ({ view: ref('tune'), setView: vi.fn() }),
}))
vi.mock('~/composables/useRecord', () => ({
  useRecord: () => ({
    sessionId: ref('session'), armed: ref(true), readings: ref(3), recording: ref(true),
    toggleRecording: calls.record, undo: calls.undo,
  }),
}))
vi.mock('~/composables/useTuner', () => ({
  AUTO_NOTE: -1,
  useTuner: () => ({
    manualNote: ref(60), setAuto: vi.fn(), stepNote: calls.step,
    frozen: ref(false), setFrozen: calls.freeze, canFreeze: ref(true), takeNow: calls.take,
  }),
}))

const probe = defineComponent({
  setup() {
    useShortcuts()
    return () => h('div', [h('button'), h('input'), h('select'), h('div', { contenteditable: 'true' }, [h('span')])])
  },
})

describe('global shortcut event ownership', () => {
  let wrapper: ReturnType<typeof mount>
  beforeEach(() => {
    vi.clearAllMocks()
    wrapper = mount(probe, { attachTo: document.body })
  })
  afterEach(() => {
    wrapper.unmount()
    document.body.replaceChildren()
  })

  function press(key: string, target: Element = wrapper.element, options: KeyboardEventInit = {}) {
    const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...options })
    target.dispatchEvent(event)
    return event
  }

  it('keeps Enter and Space available to a focused button', () => {
    const button = wrapper.get('button').element
    expect(press('Enter', button).defaultPrevented).toBe(false)
    expect(press(' ', button).defaultPrevented).toBe(false)
    expect(calls.take).not.toHaveBeenCalled()
    expect(calls.freeze).not.toHaveBeenCalled()
  })

  it('does not change recordings behind a dialog or menu', () => {
    const overlay = document.createElement('div')
    overlay.className = 'v-overlay v-overlay--active v-dialog'
    document.body.append(overlay)
    expect(press('z').defaultPrevented).toBe(false)
    press('Enter')
    press('r')
    expect(calls.undo).not.toHaveBeenCalled()
    expect(calls.take).not.toHaveBeenCalled()
    expect(calls.record).not.toHaveBeenCalled()
  })

  it('allows a shortcut while only a tooltip is visible', () => {
    const tooltip = document.createElement('div')
    tooltip.className = 'v-overlay v-overlay--active v-tooltip'
    document.body.append(tooltip)
    expect(press('z').defaultPrevented).toBe(true)
    expect(calls.undo).toHaveBeenCalledOnce()
  })

  it('leaves editable and select controls alone', () => {
    for (const selector of ['input', 'select', '[contenteditable] span']) {
      expect(press('z', wrapper.get(selector).element).defaultPrevented).toBe(false)
    }
    expect(calls.undo).not.toHaveBeenCalled()
  })

  it('ignores handled events and composition', () => {
    const event = new KeyboardEvent('keydown', { key: 'z', bubbles: true, cancelable: true })
    event.preventDefault()
    wrapper.element.dispatchEvent(event)
    press('z', wrapper.element, { isComposing: true })
    expect(calls.undo).not.toHaveBeenCalled()
  })

  it('repeats note steps but never undo or record toggles', () => {
    press('z')
    press('z', wrapper.element, { repeat: true })
    press('r', wrapper.element, { repeat: true })
    press('ArrowRight', wrapper.element, { repeat: true })
    expect(calls.undo).toHaveBeenCalledOnce()
    expect(calls.record).not.toHaveBeenCalled()
    expect(calls.step).toHaveBeenCalledWith(1)
  })

  it('removes the global handler on unmount', () => {
    wrapper.unmount()
    press('z', document.body)
    expect(calls.undo).not.toHaveBeenCalled()
  })
})
