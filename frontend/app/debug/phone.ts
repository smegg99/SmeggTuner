// The phone layout of a debug build: the window takes a phone's size, so the layout the Android app shows can be
// tested on the desktop. The choice is kept in the WebView's storage and outlives the restarts of `just dev`.

import { Window } from '@wailsio/runtime'
import { ref } from 'vue'

export const PHONE_LAYOUTS = [
  { value: 'off', labelKey: 'settings.phoneLayout.off' },
  { value: 'portrait', labelKey: 'settings.phoneLayout.portrait' },
  { value: 'landscape', labelKey: 'settings.phoneLayout.landscape' },
] as const
export type PhoneLayout = typeof PHONE_LAYOUTS[number]['value']

interface Size {
  width: number
  height: number
}

interface Stored {
  layout: PhoneLayout
  /** The window's size before the phone layout, put back when it is switched off. */
  desktop: Size
}

const STORE = 'smeggtuner.debug.phoneLayout'
// A current Android phone held upright, in CSS pixels.
const PHONE: Size = { width: 412, height: 915 }
// The narrowest phone the layout is made for; the window can be dragged down to it, upright or on its side.
const PHONE_MIN: Size = { width: 360, height: 360 }
// The main window's minimum and first size, as in backend/internal/app/window.go.
const DESKTOP_MIN: Size = { width: 960, height: 640 }
const DESKTOP: Size = { width: 1100, height: 720 }

function isDesktopSize(size: Partial<Size> | undefined): size is Size {
  return typeof size?.width === 'number' && typeof size.height === 'number'
    && size.width >= DESKTOP_MIN.width && size.height >= DESKTOP_MIN.height
}

// A missing or broken entry is the desktop window.
function load(): Stored {
  const fallback: Stored = { layout: 'off', desktop: DESKTOP }
  try {
    const stored = JSON.parse(localStorage.getItem(STORE) ?? '') as Partial<Stored>
    const known = PHONE_LAYOUTS.find(layout => layout.value === stored.layout)
    if (!known) return fallback
    return { layout: known.value, desktop: isDesktopSize(stored.desktop) ? stored.desktop : DESKTOP }
  }
  catch {
    return fallback
  }
}

const saved = load()
let desktop = saved.desktop
export const phoneLayout = ref<PhoneLayout>(saved.layout)

function save() {
  try {
    localStorage.setItem(STORE, JSON.stringify({ layout: phoneLayout.value, desktop }))
  }
  catch { /* kept for this run only */ }
}

async function resize(min: Size, size: Size) {
  await Window.UnMaximise()
  await Window.SetMinSize(min.width, min.height)
  await Window.SetSize(size.width, size.height)
}

async function apply() {
  if (phoneLayout.value === 'off') await resize(DESKTOP_MIN, desktop)
  else if (phoneLayout.value === 'portrait') await resize(PHONE_MIN, PHONE)
  else await resize(PHONE_MIN, { width: PHONE.height, height: PHONE.width })
}

export async function setPhoneLayout(next: PhoneLayout) {
  if (next === phoneLayout.value) return
  if (phoneLayout.value === 'off') {
    const size = await Window.Size()
    desktop = isDesktopSize(size) ? size : DESKTOP
  }
  phoneLayout.value = next
  save()
  await apply()
}

/** Sizes a window that has just opened for the layout chosen in an earlier run. */
export async function restorePhoneLayout() {
  if (phoneLayout.value !== 'off') await apply()
}
