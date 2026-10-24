// Which element a pointer position asks a tip of, and what the tip says; pure DOM, so it is testable.

/** Elements that carry a tip: anything titled, and every button whether titled or not. */
const TIPPED = '[title], [data-tip], button, .v-btn, [role="button"], [role="tab"]'

/** A disabled control takes no pointer events, so its parent is the target and it is found under the point. */
const DISABLED = 'button:disabled, .v-btn--disabled, [aria-disabled="true"]'

function contains(el: Element, x: number, y: number): boolean {
  const r = el.getBoundingClientRect()
  return x >= r.left && x <= r.right && y >= r.top && y <= r.bottom
}

/** tipHost is the control under (x, y) given the event target, or null; list rows are a menu's own. */
export function tipHost(target: EventTarget | null, x: number, y: number): HTMLElement | null {
  if (!(target instanceof Element)) return null
  let el = target.closest<HTMLElement>(TIPPED)
  if (!el) {
    el = [...target.querySelectorAll<HTMLElement>(DISABLED)].find(d => contains(d, x, y)) ?? null
  }
  if (!el || el.closest('.v-list-item, .v-tooltip')) return null
  return el
}

/** tipText moves a native title into data-tip, so the browser's own tooltip never shows, and names
 *  an untitled icon key by it; an untitled button falls back to its accessible name or its label. */
export function tipText(el: HTMLElement): string {
  const title = el.getAttribute('title')
  if (title !== null) {
    el.removeAttribute('title')
    el.dataset.tip = title
    const text = el.textContent?.trim() ?? ''
    if (!el.getAttribute('aria-label') && !text) el.setAttribute('aria-label', title)
  }
  return (el.dataset.tip || el.getAttribute('aria-label') || el.textContent || '').trim()
}
