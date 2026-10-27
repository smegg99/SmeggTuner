// build/android-support/import-check.mjs
import assert from 'node:assert/strict'
import { setTimeout as delay } from 'node:timers/promises'

export async function checkImport({ evaluate, binding, call, adb, uiXML, tap, until, record, filename, sessionID }) {
  await evaluate(`window.__androidCheck.opened = null;
    window.__androidCheck.call(${binding('session', 'OpenFileDialog')}, ['Android acceptance', 'Session'])
    .then(value => { window.__androidCheck.opened = { value } }, error => { window.__androidCheck.opened = { error: error.message } }); void 0`)
  await until(() => {
    const nodes = uiXML().match(/<node\b[^>]*>/g) ?? []
    const file = nodes.find(node => node.includes(`text="${filename}"`))
    if (file) { tap(file); return true }
    const downloads = nodes.find(node => node.includes('text="Downloads"') && node.includes('resource-id="android:id/title"'))
    const roots = nodes.find(node => node.includes('content-desc="Show roots"'))
    if (downloads) tap(downloads)
    else if (roots) tap(roots)
    return false
  }, 'Exported session did not appear in the native picker')
  await delay(500)
  if (!await evaluate('window.__androidCheck.opened')) {
    const resumed = adb('shell', 'dumpsys', 'activity', 'activities').split('\n').filter(line => /ResumedActivity|FocusedActivity/.test(line)).join('\n')
    if (/documentsui/i.test(resumed)) {
      adb('shell', 'input', 'keyevent', 'KEYCODE_DPAD_DOWN')
      adb('shell', 'input', 'keyevent', 'KEYCODE_ENTER')
    }
  }
  const selected = await until(() => evaluate('window.__androidCheck.opened'), 'Native import picker did not return')
  assert.match(selected.value ?? '', /\/cache\/wails-picker\//)
  assert.equal(selected.value.split('/').at(-1), filename, 'Native picker returned a different file')
  let imported
  try {
    imported = await call('session', 'ImportSession', selected.value)
    assert(imported.id && imported.id !== sessionID, 'Import must create a new session identity')
    await call('session', 'Open', imported.id)
    const snapshot = await call('session', 'Snapshot')
    assert.equal(snapshot.takes.length, 1)
    assert.equal(snapshot.takes[0].note, 69)
    record('native session picker imports the exported archive with a new identity')
  } finally {
    if (imported?.id) await call('session', 'Delete', imported.id)
    await call('session', 'Open', sessionID)
  }
}
