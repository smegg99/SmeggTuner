// build/android-support/document-check.mjs
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { writeFileSync } from 'node:fs'
import path from 'node:path'
import { checkImport } from './import-check.mjs'

export async function checkDocuments({ evaluate, binding, call, adb, serial, output, until, record,
  screenshot, uiXML, uiNode, tap, documentSave, launch, session, sessionID }) {
  await evaluate(`window.__androidCheck.picker = null;
    window.__androidCheck.call(${binding('session', 'SaveFileDialog')}, ['session', 'android-check.stsf', 'Android acceptance', 'Session'])
    .then(value => { window.__androidCheck.picker = { value } }, error => { window.__androidCheck.picker = { error: error.message } }); void 0`)
  await until(() => /documentsui|ChooserActivity/i.test(adb('shell', 'dumpsys', 'activity', 'activities').split('\n').filter(line => /ResumedActivity|FocusedActivity/.test(line)).join('\n')), 'Document picker did not open')
  await documentSave()
  screenshot('document-picker')
  if (/mInputShown=true/.test(adb('shell', 'dumpsys', 'input_method'))) {
    adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
    await until(() => !/mInputShown=true/.test(adb('shell', 'dumpsys', 'input_method')), 'Picker keyboard did not close')
  }
  adb('shell', 'input', 'keyevent', 'KEYCODE_BACK')
  const cancelled = await until(() => evaluate('window.__androidCheck.picker'), 'Document picker did not complete cancellation')
  assert.deepEqual(cancelled, { value: '' })
  record('native document picker cancellation')
  const filename = `smeggtuner-check-${Date.now()}.stsf`
  await evaluate(`window.__androidCheck.picker = null;
    window.__androidCheck.call(${binding('session', 'SaveFileDialog')}, ['session', ${JSON.stringify(filename)}, 'Android acceptance', 'Session'])
    .then(value => { window.__androidCheck.picker = { value } }, error => { window.__androidCheck.picker = { error: error.message } }); void 0`)
  const save = await documentSave()
  tap(save)
  const chosen = await until(() => evaluate('window.__androidCheck.picker'), 'Document selection did not complete')
  assert.match(chosen.value ?? '', /^content:\/\//)
  await call('session', 'ExportSession', sessionID, chosen.value)
  const exported = execFileSync('adb', ['-s', serial, 'exec-out', 'cat', `/sdcard/Download/${filename}`], { timeout: 15000 })
  assert.equal(exported.subarray(0, 2).toString(), 'PK', 'Export did not create a session archive')
  const archive = path.join(output, filename)
  writeFileSync(archive, exported)
  const saved = JSON.parse(execFileSync('python3', ['-c', 'import json,sys,zipfile; print(zipfile.ZipFile(sys.argv[1]).read("session.json").decode())', archive]).toString())
  assert.equal(saved.id, sessionID)
  assert.equal(saved.takes.length, 1)
  record('native document export contains the saved session')
  await checkImport({ evaluate, binding, call, adb, uiXML, tap, until, record, filename, sessionID })
  const printing = await call('report', 'Export', { format: 'pdf', date: '' })
  assert.equal(printing.printing, true)
  assert.equal(printing.path, '')
  await until(() => /printspooler/i.test(adb('shell', 'dumpsys', 'activity', 'activities').split('\n').filter(line => /ResumedActivity|FocusedActivity/.test(line)).join('\n')), 'Android print UI did not open')
  await until(() => uiNode('com.android.printspooler:id/preview_page'), 'Print preview did not finish rendering')
  screenshot('print-preview')
  record('native print UI accepted report')
  tap(await until(() => uiNode('com.android.printspooler:id/destination_spinner'), 'Print destination selector did not render'))
  const pdfDestination = await until(() => uiXML().match(/<node\b[^>]*>/g)?.find(node => node.includes('text="Save as PDF"')), 'Save as PDF destination is missing')
  tap(pdfDestination)
  tap(await until(() => uiNode('com.android.printspooler:id/print_button'), 'Save as PDF action did not become available'))
  const pdfSave = await documentSave()
  const nameField = uiXML().match(/<node\b[^>]*>/g)?.find(node => node.includes('class="android.widget.EditText"'))
  const pdfName = /text="([^"]+\.pdf)"/.exec(nameField ?? '')?.[1]
  assert.match(pdfName ?? '', /^android-acceptance-.*\.pdf$/)
  assert(pdfName.includes(session.name.toLowerCase().replaceAll(' ', '-')), 'PDF filename does not belong to this check session')
  adb('shell', 'test', '!', '-e', `/sdcard/Download/${pdfName}`)
  tap(pdfSave)
  const pdf = await until(() => {
    try {
      const data = execFileSync('adb', ['-s', serial, 'exec-out', 'cat', `/sdcard/Download/${pdfName}`], { timeout: 15000, stdio: ['ignore', 'pipe', 'pipe'] })
      return data.subarray(0, 5).toString() === '%PDF-' && data
    } catch { return false }
  }, 'Android did not save a PDF')
  writeFileSync(path.join(output, pdfName), pdf)
  record('system Save as PDF created a PDF file')
  launch()
}
