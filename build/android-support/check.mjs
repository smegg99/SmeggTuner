// build/android-support/check.mjs
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { setTimeout as delay } from 'node:timers/promises'
import { checkNavigation } from './navigation-check.mjs'
import { checkDocuments } from './document-check.mjs'
import { checkLayout } from './layout-check.mjs'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..')
const config = JSON.parse(readFileSync(path.join(root, 'build/android-support/config.json'), 'utf8'))
const serial = process.env.ANDROID_SERIAL
const profile = process.argv[2] ?? 'emulator'
assert(['emulator', 'phone'].includes(profile), 'Check profile must be emulator or phone')
assert(serial && !/\s/.test(serial), 'Set ANDROID_SERIAL to an authorized device')
if (profile === 'emulator') assert.match(serial, /^emulator-\d+$/, 'Set ANDROID_SERIAL to a dedicated test emulator')
const output = path.join(root, 'bin/android/checks', `${serial.replace(/[^a-zA-Z0-9_.-]/g, '_')}-${Date.now()}`)
mkdirSync(output, { recursive: true })
const adb = (...args) => execFileSync('adb', ['-s', serial, ...args], { timeout: 30000 }).toString().trim()
const report = { serial, profile, api: adb('shell', 'getprop', 'ro.build.version.sdk'),
  emulator: adb('shell', 'getprop', 'ro.kernel.qemu') === '1', checks: [] }
if (profile === 'phone') report.manualChecks = ['Pitch accuracy with a known reference tone',
  'Native Back, rotation and keyboard', 'Document import/export and Save as PDF']
const record = (name, details = true) => {
  report.checks.push({ name, details })
  console.log(`PASS ${name}`)
}
async function until(check, message, timeout = 15000) {
  const end = Date.now() + timeout
  while (Date.now() < end) {
    const result = await check()
    if (result) return result
    await delay(250)
  }
  throw new Error(message)
}
function binding(service, method) {
  const source = readFileSync(path.join(root, 'frontend/bindings/smegg.me/smeggtuner/backend/services', service, 'service.ts'), 'utf8')
  const body = source.split(`export function ${method}(`)[1]
  assert(body, `Missing generated binding ${service}.${method}`)
  const id = /\$Call.ByID\((\d+)/.exec(body.split('\n}')[0])
  assert(id, `Missing method ID for ${service}.${method}`)
  return Number(id[1])
}
const processRow = adb('shell', 'ps').split('\n').map(row => row.trim().split(/\s+/)).find(row => row.at(-1) === config.applicationId)
const pid = processRow?.[1] ?? ''
assert.match(pid, /^\d+$/, 'The application must be running')
const port = adb('forward', 'tcp:0', `localabstract:webview_devtools_remote_${pid}`)
let socket
let sequence = 0
const pending = new Map()
const errors = []
let previousSession
let sessionID
let bridgeInstalled = false
function cdp(method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = ++sequence
    const timer = setTimeout(() => { pending.delete(id); reject(new Error(`${method} timed out`)) }, 20000)
    pending.set(id, { resolve, reject, timer })
    socket.send(JSON.stringify({ id, method, params }))
  })
}
async function evaluate(expression) {
  const result = await cdp('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true })
  assert(!result.exceptionDetails, JSON.stringify(result.exceptionDetails))
  return result.result.value
}
const call = (service, method, ...args) => evaluate(`window.__androidCheck.call(${binding(service, method)}, ${JSON.stringify(args)})`)
function screenshot(name) {
  writeFileSync(path.join(output, `${name}.png`), execFileSync('adb', ['-s', serial, 'exec-out', 'screencap', '-p'], { timeout: 15000 }))
}
function launch() {
  adb('shell', 'am', 'start', '-W', '-n', `${config.applicationId}/com.wails.app.LauncherActivity`)
}
function uiXML() {
  const dump = '/data/local/tmp/smeggtuner-acceptance.xml'
  let failure
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      const result = adb('shell', 'uiautomator', 'dump', dump)
      if (!result.includes('dumped to:')) { failure = new Error(result); continue }
      const xml = adb('shell', 'cat', dump)
      adb('shell', 'rm', '-f', dump)
      writeFileSync(path.join(output, 'native-ui.xml'), xml)
      return xml
    } catch (error) { failure = error }
  }
  throw failure
}
function uiNode(id) {
  return uiXML().match(/<node\b[^>]*>/g)?.find(node => node.includes(`resource-id="${id}"`) && node.includes('enabled="true"'))
}
function tap(node) {
  const bounds = /bounds="\[(\d+),(\d+)\]\[(\d+),(\d+)\]"/.exec(node)
  assert(bounds, 'Native control has no bounds')
  adb('shell', 'input', 'tap', String(Math.floor((Number(bounds[1]) + Number(bounds[3])) / 2)), String(Math.floor((Number(bounds[2]) + Number(bounds[4])) / 2)))
}
async function documentSave() {
  return until(() => {
    const nodes = uiXML().match(/<node\b[^>]*>/g) ?? []
    const downloads = nodes.find(node => node.includes('text="Downloads"') && node.includes('resource-id="android:id/title"'))
    const save = nodes.find(node => node.includes('resource-id="android:id/button1"') && node.includes('enabled="true"'))
    if (save && !downloads) return save
    if (downloads) tap(downloads)
    else {
      const roots = nodes.find(node => node.includes('content-desc="Show roots"'))
      if (roots) tap(roots)
    }
    return false
  }, 'Document picker has no writable Downloads destination')
}
try {
  const targets = await until(async () => {
    const list = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json()
    return list.find(target => target.type === 'page' && target.url.startsWith('https://wails.localhost'))
  }, 'Application WebView was not created')
  socket = new WebSocket(targets.webSocketDebuggerUrl)
  await new Promise((resolve, reject) => { socket.onopen = resolve; socket.onerror = reject })
  socket.onmessage = ({ data }) => {
    const response = JSON.parse(data)
    if (response.method === 'Runtime.exceptionThrown') errors.push(response.params)
    const request = pending.get(response.id)
    if (!request) return
    pending.delete(response.id)
    clearTimeout(request.timer)
    if (response.error) request.reject(new Error(JSON.stringify(response.error)))
    else request.resolve(response.result)
  }
  await cdp('Runtime.enable')
  await until(() => evaluate('document.querySelectorAll("canvas").length === 4 && Boolean(document.querySelector(".v-application"))'), 'Frontend did not render')
  const appearance = await evaluate(`({
    provider: navigator.userAgent,
    theme: getComputedStyle(document.documentElement).getPropertyValue('--v-theme-background').trim(),
    width: innerWidth, height: innerHeight,
    error: document.title.includes('500') || document.body.innerText.includes('Internal Server Error'),
    fonts: [...document.fonts].filter(font => font.status === 'loaded').map(font => font.family)
  })`)
  assert(!appearance.error && appearance.theme, 'Frontend has no usable theme')
  record('frontend and theme', appearance)
  screenshot('tuner')
  await evaluate(`(() => {
    const previousCallback = window._wailsAndroidCallback
    const previousEvent = window._wails.dispatchWailsEvent
    const requests = new Map()
    const state = { measurements: 0, states: [] }
    window._wailsAndroidCallback = (id, reply, error) => {
      const request = requests.get(id)
      if (!request) return previousCallback?.(id, reply, error)
      requests.delete(id)
      clearTimeout(request.timer)
      try {
        const result = JSON.parse(reply || '{}')
        if (error || !result.ok) request.reject(new Error(error || JSON.stringify(result.error)))
        else request.resolve('data' in result ? result.data : result.text)
      } catch (failure) { request.reject(failure) }
    }
    window._wails.dispatchWailsEvent = function(event) {
      if (event.name === 'tuner:measurement') state.measurements++
      if (event.name === 'tuner:state') state.states.push(event.data)
      return previousEvent(event)
    }
    state.call = (methodID, args) => new Promise((resolve, reject) => {
      const id = 'android-check-' + Math.random()
      const timer = setTimeout(() => { requests.delete(id); reject(new Error('Binding timed out')) }, 90000)
      requests.set(id, { resolve, reject, timer })
      window.wails.invokeAsync(id, JSON.stringify({ object: 0, method: 0,
        clientId: window._wails.clientId, args: { 'call-id': id, methodID, args } }))
    })
    state.dispose = () => {
      window._wailsAndroidCallback = previousCallback
      window._wails.dispatchWailsEvent = previousEvent
      for (const request of requests.values()) { clearTimeout(request.timer); request.reject(new Error('Check ended')) }
      delete window.__androidCheck
    }
    window.__androidCheck = state
  })()`)
  bridgeInstalled = true
  const settings = await call('config', 'GetConfig')
  assert(settings && typeof settings === 'object')
  record('generated config binding')
  await call('tuner', 'Start')
  await until(() => evaluate('window.__androidCheck.measurements >= 5'), 'No live measurement events')
  assert.equal(await call('tuner', 'IsRunning'), true)
  record('microphone capture and frontend events', await evaluate('window.__androidCheck.measurements'))
  adb('shell', 'input', 'keyevent', 'KEYCODE_HOME')
  await until(async () => (await call('tuner', 'IsRunning')) === false, 'Capture continued in background')
  record('background stops capture')
  launch()
  assert.equal(await call('tuner', 'IsRunning'), false, 'Resume must not silently restart capture')
  record('resume leaves capture stopped')
  previousSession = await call('session', 'Active')
  if (previousSession) await call('session', 'Save')
  const session = await call('session', 'Create', { name: `Android acceptance ${Date.now()}`, instrument: { reedCount: 1, a4: 440, lo: 69, hi: 71 } })
  sessionID = session.id
  assert(sessionID)
  await call('session', 'UpsertTake', { note: 69, at: new Date().toISOString(), reeds: [{ present: true, freq: 440, dev: 0 }] })
  await call('session', 'Save')
  await call('session', 'Close')
  const reopened = await call('session', 'Open', sessionID)
  assert.equal(reopened.id, sessionID)
  assert.equal((await call('session', 'Snapshot')).takes.length, 1)
  record('session save, close and reopen')
  if (profile === 'emulator') {
    await checkNavigation({ evaluate, call, adb, until, record, sessionID })
    await checkDocuments({ evaluate, binding, call, adb, serial, output, until, record,
      screenshot, uiXML, uiNode, tap, documentSave, launch, session, sessionID })
    await checkLayout({ evaluate, adb, uiXML, until, screenshot, record })
  }
  await call('session', 'Delete', sessionID)
  sessionID = undefined
  if (previousSession) await call('session', 'Open', previousSession.id)
  screenshot('completed')
  assert.deepEqual(errors, [], 'WebView raised an uncaught exception')
  record('no uncaught WebView exceptions')
} catch (error) {
  report.error = error.stack
  try { report.captureStates = await evaluate('window.__androidCheck?.states') } catch { /* Preserve the original failure. */ }
  try { screenshot('failure') } catch { /* Preserve the original failure. */ }
  throw error
} finally {
  try {
    if (bridgeInstalled && socket?.readyState === WebSocket.OPEN) {
      await call('tuner', 'Stop')
      if (sessionID) { launch(); await call('session', 'Delete', sessionID) }
      if (sessionID && previousSession) await call('session', 'Open', previousSession.id)
      await evaluate('window.__androidCheck?.dispose()')
    }
  } catch (error) {
    report.cleanupError = error.stack
    process.exitCode = 1
  } finally {
    socket?.close()
    for (const request of pending.values()) clearTimeout(request.timer)
    adb('forward', '--remove', `tcp:${port}`)
    writeFileSync(path.join(output, 'result.json'), JSON.stringify(report, null, 2) + '\n')
    try {
      const args = ['logcat', '-d', '-t', '1500', '-v', 'threadtime']
      if (profile === 'phone' && Number(report.api) >= 24) args.push(`--pid=${pid}`)
      const logs = adb(...args)
      writeFileSync(path.join(output, 'logcat.txt'), profile === 'phone'
        ? logs.split('\n').filter(line => line.trim().split(/\s+/)[2] === pid).join('\n') : logs)
    } catch { /* Device may have stopped. */ }
    console.log(`Android ${profile} evidence: ${output}`)
    if (report.manualChecks) console.log(`Manual checks: ${report.manualChecks.join('; ')}`)
  }
}
