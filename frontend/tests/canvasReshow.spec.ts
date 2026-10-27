import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { useCanvas } from '~/composables/useCanvas'

// WebKitGTK drops a canvas's pixels while the tuner view is hidden (v-show) and does not give them back,
// so a canvas shown again at its old size has to paint again by itself, not wait for a pointer to wake it.

const size = { width: 0, height: 0 }
let resize: () => void = () => {}
let frames: FrameRequestCallback[] = []
let now = 0

function flush() {
  for (let i = 0; i < 5 && frames.length; i++) {
    const run = frames
    frames = []
    now += 100
    run.forEach(f => f(now))
  }
}

function show(width: number, height: number) {
  size.width = width
  size.height = height
  resize()
  flush()
}

function mountCanvas() {
  const background = vi.fn()
  const Probe = defineComponent({
    setup() {
      const { host, canvas } = useCanvas({ draw: () => {}, background })
      return () => h('div', { ref: host }, [h('canvas', { ref: canvas })])
    },
  })
  mount(Probe, { attachTo: document.body })
  flush()
  return background
}

beforeEach(() => {
  frames = []
  vi.stubGlobal('requestAnimationFrame', (f: FrameRequestCallback) => frames.push(f))
  vi.stubGlobal('cancelAnimationFrame', () => {})
  vi.stubGlobal('ResizeObserver', class {
    constructor(callback: () => void) { resize = callback }
    observe() {}
    disconnect() {}
  })
  vi.spyOn(HTMLElement.prototype, 'clientWidth', 'get').mockImplementation(() => size.width)
  vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockImplementation(() => size.height)
  vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(() => ({
    setTransform() {}, clearRect() {}, drawImage() {},
  }) as unknown as CanvasRenderingContext2D)
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('a canvas panel', () => {
  it('paints again when it is shown at the size it was hidden at', () => {
    const background = mountCanvas()
    show(300, 100)
    expect(background).toHaveBeenCalledTimes(1)

    show(0, 0)
    show(300, 100)
    expect(background).toHaveBeenCalledTimes(2)
  })

  it('paints again when the window comes back from the tray or a minimise', () => {
    const background = mountCanvas()
    show(300, 100)
    expect(background).toHaveBeenCalledTimes(1)

    document.dispatchEvent(new Event('visibilitychange'))
    flush()
    expect(background).toHaveBeenCalledTimes(2)
  })
})
