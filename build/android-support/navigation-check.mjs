// build/android-support/navigation-check.mjs
import assert from 'node:assert/strict'

export async function checkNavigation({ evaluate, call, adb, until, record, sessionID }) {
  const session = await call('session', 'Active')
  assert.equal(session.id, sessionID)
  const click = async (selector, text) => evaluate(`(() => {
    const button = [...document.querySelectorAll(${JSON.stringify(selector)})]
      .find(element => element.title === ${JSON.stringify(text)} || element.textContent.trim() === ${JSON.stringify(text)})
    if (!button) throw new Error('Navigation control is missing: ' + ${JSON.stringify(text)})
    button.click()
  })()`)
  const shown = selector => evaluate(`Boolean(document.querySelector(${JSON.stringify(selector)}))`)
  const back = () => adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
  const openSession = async () => {
    await until(() => evaluate(`Boolean([...document.querySelectorAll('.card__open')].find(element => element.textContent.includes(${JSON.stringify(session.name)})))`), 'Saved session card did not render')
    await evaluate(`[...document.querySelectorAll('.card__open')].find(element => element.textContent.includes(${JSON.stringify(session.name)})).click()`)
    await until(() => shown('.read'), 'Session detail did not open')
  }
  await click('.bar__views button, .window__navigation button', 'Settings')
  await until(() => shown('.settings'), 'Settings did not open')
  await click('.status__action', 'About')
  await until(() => shown('.v-dialog.v-overlay--active'), 'About dialog did not open')
  back()
  await until(async () => !await shown('.v-dialog.v-overlay--active'), 'Back did not close the top dialog')
  assert.equal(await shown('.settings'), true, 'Back navigated behind the dialog')
  record('native Back closes only the top overlay')
  back()
  await until(async () => !await shown('.settings'), 'Back did not return from Settings')
  record('native Back returns from Settings')
  // The workshop is hidden in 1.4 (utils/features.ts); its session and calibration flows are checked only when shown.
  if (!await evaluate(`[...document.querySelectorAll('.bar__views button, .window__navigation button')].some(button => button.title === 'Workshop')`)) {
    record('workshop hidden in this build: session and calibration navigation not reachable')
    return
  }

  await click('.bar__views button, .window__navigation button', 'Workshop')
  await openSession()
  await click('.ident__views button', 'Accordion')
  await until(() => shown('.acc__form input'), 'Accordion editor did not open')
  // Edit and Back immediately: the 600 ms debounce must be flushed before unmount.
  const edited = `Android Back ${Date.now()}`
  await evaluate(`(() => {
    const input = document.querySelector('.acc__form input')
    input.value = ${JSON.stringify(edited)}
    input.dispatchEvent(new Event('input', { bubbles: true }))
  })()`)
  back()
  await until(async () => !await shown('.acc') && await shown('.read'), 'Back did not return from the editor tab')
  assert.equal((await call('session', 'Snapshot')).instrument.name, edited)
  record('native Back flushes pending accordion edits before leaving the tab')

  await click('.ident__views button', 'Accordion')
  await until(() => shown('.acc__form input'), 'Accordion editor did not reopen')
  await evaluate(`(() => {
    const input = document.querySelector('.acc__form input')
    input.value = ''
    input.dispatchEvent(new Event('input', { bubbles: true }))
  })()`)
  back()
  await until(() => shown('.status__error'), 'Invalid draft did not show a navigation error')
  assert.equal(await shown('.acc'), true, 'Back discarded an invalid draft')
  await evaluate(`(() => {
    const input = document.querySelector('.acc__form input')
    input.value = ${JSON.stringify(edited)}
    input.dispatchEvent(new Event('input', { bubbles: true }))
  })()`)
  back()
  await until(async () => !await shown('.acc') && await shown('.read'), 'Corrected draft could not be left')
  record('native Back retains a refused draft and allows retry')

  await click('button', 'Calibrate')
  await until(() => shown('.cal'), 'Calibration did not open')
  await until(() => shown('.sw--setup'), 'Calibration sweep setup did not open')
  await click('.sw--setup button', 'Start the sweep')
  await until(async () => (await call('record', 'State')).armed && await shown('.sw__ask'), 'Calibration sweep did not arm')
  back()
  await until(() => shown('.sw--setup'), 'Back did not return the sweep to setup')
  assert.equal((await call('record', 'State')).armed, false)
  record('native Back disarms the sweep and returns to setup')
  await click('.sw--setup button', 'Start the sweep')
  await until(() => shown('.sw__ask'), 'Calibration sweep did not restart')
  await click('.sw__keys button', 'Stop')
  await until(() => shown('.sw--done'), 'Stop did not finish the sweep')
  assert.equal((await call('record', 'State')).armed, false)
  record('calibration Stop disarms recording')
  back()
  await until(() => shown('.sw--setup'), 'Back did not leave the completed sweep')
  back()
  await until(() => shown('.read'), 'Back did not leave calibration')
  assert.equal((await call('record', 'State')).armed, false)
  record('native Back disarms calibration before returning to the session')
  back()
  await until(async () => (await call('session', 'Active')) === null, 'Back did not close the session')
  await until(() => shown('.card__open'), 'Workshop list did not return')
  record('native Back saves and closes the session')
  back()
  await until(async () => !await shown('.shop'), 'Back did not return to the tuner')
  await call('session', 'Open', sessionID)
}
