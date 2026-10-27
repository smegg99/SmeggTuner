import { afterEach, describe, expect, it } from 'vitest'
import { tipHost, tipText } from '~/utils/tooltip'

// The app's one tooltip: which control the pointer is on, and what its tip says.

function place(html: string) {
  document.body.innerHTML = html
}

// happy-dom lays nothing out, so each test gives the elements the boxes a browser would.
function box(el: Element, left: number, top: number, width: number, height: number) {
  el.getBoundingClientRect = () => ({ left, top, right: left + width, bottom: top + height, width, height, x: left, y: top, toJSON: () => ({}) })
}

afterEach(() => {
  document.body.innerHTML = ''
})

describe('the control under the pointer', () => {
  it('is the titled button around the target, not the icon inside it', () => {
    place('<button title="Stop"><i id="icon"></i></button>')
    expect(tipHost(document.getElementById('icon'), 0, 0)).toBe(document.querySelector('button'))
  })

  it('is a disabled button under the point when the event lands on its parent', () => {
    place('<div id="row"><button id="a" title="One">1</button><button id="b" title="Two" disabled>2</button></div>')
    box(document.getElementById('a')!, 0, 0, 30, 30)
    box(document.getElementById('b')!, 40, 0, 30, 30)
    const row = document.getElementById('row')
    expect(tipHost(row, 50, 10)?.id).toBe('b')
    expect(tipHost(row, 100, 10)).toBeNull()
  })

  it('leaves a menu row to its menu', () => {
    place('<div class="v-list-item"><button id="row">Row</button></div>')
    expect(tipHost(document.getElementById('row'), 0, 0)).toBeNull()
  })
})

describe('what a tip says', () => {
  it('takes the native title over, so the browser shows no second tooltip', () => {
    place('<button id="k"></button>')
    const k = document.getElementById('k')!
    k.setAttribute('title', 'Hold the reading')
    expect(tipText(k)).toBe('Hold the reading')
    expect(k.hasAttribute('title')).toBe(false)
    expect(k.getAttribute('aria-label')).toBe('Hold the reading')
  })

  it('follows a title the app changes later', () => {
    place('<button id="k" title="Start"></button>')
    const k = document.getElementById('k')!
    tipText(k)
    k.setAttribute('title', 'Stop')
    expect(tipText(k)).toBe('Stop')
  })

  it('falls back to the accessible name, then the label, of an untitled button', () => {
    place('<button id="a" aria-label="Volume"></button><button id="b"> Cancel </button>')
    expect(tipText(document.getElementById('a')!)).toBe('Volume')
    expect(tipText(document.getElementById('b')!)).toBe('Cancel')
  })
})
