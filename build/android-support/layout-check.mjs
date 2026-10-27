// build/android-support/layout-check.mjs
import assert from 'node:assert/strict'
import { setTimeout as delay } from 'node:timers/promises'

export async function checkLayout({ evaluate, adb, uiXML, until, screenshot, record }) {
  const rotation = adb('shell', 'settings', 'get', 'system', 'user_rotation')
  const automatic = adb('shell', 'settings', 'get', 'system', 'accelerometer_rotation')
  const click = title => evaluate(`(() => {
    const button = [...document.querySelectorAll('button')].find(button => button.title === ${JSON.stringify(title)})
    if (!button) throw new Error('Missing layout control: ' + ${JSON.stringify(title)})
    button.click()
  })()`)
  // The workshop is hidden in 1.4 (utils/features.ts); what it adds is checked only when the build shows it.
  const workshop = await evaluate(`[...document.querySelectorAll('button')].some(button => button.title === 'Workshop')`)
  const groups = ['Microphone', 'Detection', 'Readings', ...(workshop ? ['Workshop details'] : []), 'Appearance']
  const checkSettings = async orientation => {
    await click('Settings')
    await until(() => evaluate('Boolean(document.querySelector(".settings"))'), 'Settings did not render')
    const sections = await evaluate(`(() => {
      const body = document.querySelector('.settings__body')
      return [...body.querySelectorAll('.settings__section')].map(section => {
        const rect = section.getBoundingClientRect()
        const controls = [...section.querySelectorAll('button, input:not([type="hidden"]), textarea')]
          .filter(control => control.getBoundingClientRect().width > 0)
        const reachable = controls.every(control => {
          control.scrollIntoView({ block: 'nearest', inline: 'nearest' })
          const field = control.getBoundingClientRect(), group = section.getBoundingClientRect()
          const bounds = body.getBoundingClientRect()
          return field.top >= Math.max(bounds.top, group.top) - 1
            && field.bottom <= Math.min(bounds.bottom, group.bottom) + 1
            && field.left >= bounds.left - 1 && field.right <= bounds.right + 1
        })
        return { title: section.querySelector('.form-section__title')?.textContent.trim(),
          visible: rect.width > 0 && rect.height > 0, controls: controls.length, reachable }
      })
    })()`)
    assert.deepEqual(sections.map(section => section.title), groups)
    assert(sections.every(section => section.visible && section.controls > 0 && section.reachable),
      `${orientation} settings must expose every group and control: ${JSON.stringify(sections)}`)
    await evaluate('document.querySelector(".settings__body").scrollTop = 0')
    screenshot(`settings-${orientation}`)
    record(`${orientation} settings exposes all ${groups.length} desktop groups`, sections)
  }
  // Taps the last visible field matching selector natively; the keyboard must leave it in view.
  const checkKeyboard = async (selector, name) => {
    const input = await evaluate(`(() => {
      const field = [...document.querySelectorAll(${JSON.stringify(selector)})].filter(field => field.getBoundingClientRect().height > 0).at(-1)
      field.scrollIntoView({ block: 'center', inline: 'nearest' })
      const rect = field.getBoundingClientRect()
      return { x: rect.x + rect.width / 2, y: rect.y + rect.height / 2, width: innerWidth }
    })()`)
    const webview = uiXML().match(/<node\b[^>]*class="android.webkit.WebView"[^>]*>/)?.[0]
    const bounds = /bounds="\[(\d+),(\d+)\]\[(\d+),(\d+)\]/.exec(webview ?? '')
    assert(bounds, 'Native WebView bounds are missing')
    const left = Number(bounds[1]), top = Number(bounds[2])
    const scale = (Number(bounds[3]) - left) / input.width
    adb('shell', 'input', 'tap', String(Math.round(left + input.x * scale)), String(Math.round(top + input.y * scale)))
    await until(() => /mInputShown=true/.test(adb('shell', 'dumpsys', 'input_method')), 'Native keyboard did not open')
    await until(() => evaluate(`(() => {
      const field = document.activeElement, rect = field.getBoundingClientRect()
      return field.tagName === 'INPUT' && rect.top >= 0 && rect.bottom <= visualViewport.height
    })()`), `Keyboard obscured the focused ${name}`)
    screenshot('keyboard')
    adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
    await until(() => evaluate('Math.abs(visualViewport.offsetLeft) <= 1 && Math.abs(visualViewport.offsetTop) <= 1 && visualViewport.width >= innerWidth - 1'), 'Closing the keyboard left the desktop zoomed or panned')
    record(`native keyboard keeps the ${name} visible and restores the fitted desktop`)
  }
  try {
    await click('Tune')
    adb('shell', 'settings', 'put', 'system', 'accelerometer_rotation', '0')
    adb('shell', 'settings', 'put', 'system', 'user_rotation', '1')
    await until(() => evaluate('innerWidth > innerHeight'), 'Emulator did not rotate to landscape')
    const layout = await until(() => evaluate(`(() => {
      const views = [...document.querySelectorAll('.bar__views button, .window__navigation button')].map(button => button.getBoundingClientRect())
      return views.length === ${workshop ? 3 : 2} && views.every(rect => rect.width > 0 && rect.left >= 0 && rect.right <= innerWidth)
        && document.documentElement.scrollWidth <= innerWidth && { width: innerWidth, height: innerHeight }
    })()`), 'Landscape clipped navigation')
    assert.equal(await evaluate('document.querySelectorAll(".tuner__tabs").length'), 0, 'The desktop plots must not be replaced by mobile tabs')
    assert.equal(await evaluate('document.querySelectorAll(".window__navigation").length'), 0, 'Navigation must stay in the desktop toolbar')
    await until(() => evaluate(`(() => {
      const canvases = [...document.querySelectorAll('.tuner canvas')].map(canvas => canvas.getBoundingClientRect())
      const scrolling = [...document.querySelectorAll('.window, .window *')].some(element => {
        const rect = element.getBoundingClientRect()
        return rect.height > 0 && element.clientHeight > 0 && element.scrollHeight > element.clientHeight + 2
          && /auto|scroll/.test(getComputedStyle(element).overflowY)
      })
      return innerWidth >= 960 && innerHeight >= 640 && canvases.length === 4
        && canvases.every(rect => rect.height >= 30 && rect.width > 0 && rect.bottom <= visualViewport.height + 1)
        && !scrolling
    })()`), 'The scaled desktop must show all four plots without scrolling')
    screenshot('landscape')
    record('scaled desktop layout keeps all plots and toolbar navigation visible', layout)
    await checkSettings('landscape')
    adb('shell', 'settings', 'put', 'system', 'user_rotation', '0')
    await until(() => evaluate('innerHeight > innerWidth'), 'Emulator did not rotate to portrait')
    await checkSettings('portrait')
    if (!workshop) {
      // Settings is still open from checkSettings; its tolerance fields are always enabled text fields.
      await checkKeyboard('.settings__body .v-text-field:not(.v-select) input', 'tolerance field')
      return
    }
    await click('Workshop')
    await until(() => evaluate('Boolean(document.querySelector(".shop"))'), 'Workshop did not render after rotation')
    await click('New session')
    await until(() => evaluate('[...document.querySelectorAll(".v-dialog.v-overlay--active input")].some(field => field.getBoundingClientRect().height > 0)'), 'Desktop session dialog did not render')
    assert.equal(await evaluate('document.querySelectorAll(".v-dialog.v-overlay--active .pages").length'), 0, 'The session dialog must retain its desktop form')
    const dialog = await evaluate(`(() => {
      const overlay = document.querySelector('.v-dialog.v-overlay--active')
      const card = overlay.querySelector('.dlg'), rect = card.getBoundingClientRect()
      const head = card.querySelector('.panel__head').getBoundingClientRect()
      const body = card.querySelector('.dlg__body').getBoundingClientRect()
      const foot = card.querySelector('.panel__foot').getBoundingClientRect()
      const actions = [...card.querySelectorAll('.dlg__foot button')].map(button => button.getBoundingClientRect())
      return { named: overlay.getAttribute('aria-label') === card.querySelector('h2').textContent.trim(),
        fits: rect.left >= 0 && rect.right <= innerWidth && rect.top >= 0 && rect.bottom <= visualViewport.height,
        separated: head.bottom <= body.top + 1 && body.bottom <= foot.top + 1,
        actions: actions.length === 2 && actions.every(action => action.height >= 32 && action.top >= foot.top && action.bottom <= foot.bottom) }
    })()`)
    assert(Object.values(dialog).every(Boolean), `Dialog frame or actions are clipped: ${JSON.stringify(dialog)}`)
    record('dialog keeps its named header, complete form and full-size actions separate', dialog)
    await checkKeyboard('.v-dialog.v-overlay--active input', 'session field')
  } finally {
    if (/mInputShown=true/.test(adb('shell', 'dumpsys', 'input_method'))) {
      adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
      await delay(250)
    }
    if (await evaluate('Boolean(document.querySelector(".v-dialog.v-overlay--active"))')) {
      adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
      await until(() => evaluate('!document.querySelector(".v-dialog.v-overlay--active")'), 'Back did not dismiss the session dialog')
    }
    await click('Tune')
    for (const [key, value] of [['user_rotation', rotation], ['accelerometer_rotation', automatic]]) {
      adb('shell', 'settings', value === 'null' ? 'delete' : 'put', 'system', key, ...(value === 'null' ? [] : [value]))
    }
  }
}
