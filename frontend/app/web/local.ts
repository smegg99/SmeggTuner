// app/web/local.ts
// Bound methods the browser answers itself: the desktop audio, theme, title, browser and logger services have no Go side here.

// MicControl is the browser microphone the audio service calls stand for.
export interface MicControl {
  readonly deviceId: string
  readonly label: string
  select(deviceId: string): Promise<void>
}

type Handler = (args: unknown[]) => unknown

// bindingID is the Wails binding ID: FNV-32a of "<pkgpath>.<Type>.<Method>".
export function bindingID(fqn: string): number {
  let h = 0x811C9DC5
  for (const b of new TextEncoder().encode(fqn)) {
    h ^= b
    h = Math.imul(h, 0x01000193) >>> 0
  }
  return h
}

const AUDIO = 'smegg.me/smeggtuner/backend/services/audio.Service.'
const TUNER = 'smegg.me/smeggtuner/backend/services/tuner.Service.'
const S99 = 'github.com/smegg99/s99wails/v2/services/'

// Names from the last listing; before the microphone is opened the browser gives the current device no label of its own.
const names = new Map<string, string>()

// Browser aliases for "whatever the system uses"; the menu's own System default (the empty id) already is that.
const ALIASES = new Set(['default', 'communications'])

// Without microphone permission browsers list inputs with blank ids and names; those cannot be picked, so they are left out.
async function listMics() {
  const all = await navigator.mediaDevices?.enumerateDevices?.() ?? []
  const mics = all.filter(d => d.kind === 'audioinput' && d.deviceId && !ALIASES.has(d.deviceId))
  mics.forEach(d => names.set(d.deviceId, d.label))
  return mics.map(d => ({ id: d.deviceId, name: d.label, default: d.deviceId === 'default' }))
}

function dark() {
  return typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches
}

// save persists the picked microphone, as the desktop audio service does in its config.
export function localServices(mic: MicControl, save: (deviceId: string) => Promise<unknown> = async () => {}): Map<number, Handler> {
  const quietTransport = { available: false, name: '', duration: 0, position: 0, from: 0, to: 0, paused: false, moving: false }
  const table: Record<string, Handler> = {
    [AUDIO + 'ListDevices']: () => listMics(),
    [AUDIO + 'Current']: () => ({ kind: 'mic', deviceId: mic.deviceId, path: '', loop: false, name: mic.label || names.get(mic.deviceId) || '' }),
    [AUDIO + 'SelectMic']: async (args) => {
      const id = String(args[0] ?? '')
      await mic.select(id)
      await save(id)
      return null
    },
    [AUDIO + 'OpenFileDialog']: () => '',
    [AUDIO + 'Transport']: () => quietTransport,
    [AUDIO + 'Muted']: () => false,
    [AUDIO + 'Volume']: () => 1,
    [AUDIO + 'SetMuted']: () => false,
    [AUDIO + 'SetVolume']: args => Number(args[0] ?? 1),
    // "unreliable" means no correction is applied and the start never asks for a calibration; a browser cannot be calibrated.
    [AUDIO + 'CalibrationState']: args => ({ deviceId: String(args[0] ?? ''), status: 'unreliable', percent: 0, elapsed: 0, progress: 0, settled: false, spread: 0, rate: 0 }),
    [AUDIO + 'AutoVolume']: args => ({ deviceId: String(args[0] ?? ''), on: false, supported: false, active: false, level: 0 }),
    [AUDIO + 'Correction']: () => ({ percent: 0, missing: false }),
    // The web edition has no reference tone (the Go tone player is desktop-only); pressing a note only picks it.
    [TUNER + 'PlayTone']: () => null,
    [TUNER + 'ToggleTone']: () => null,
    [TUNER + 'StopTone']: () => null,
    [TUNER + 'SetToneVolume']: args => Number(args[0] ?? 1),
    // No accent: a browser cannot read the system's, so the theme falls back to the web default (WEB_ACCENT).
    [S99 + 'theme.Service.GetTheme']: () => ({ accent_color: '', dark_mode: dark(), scheme: dark() ? 'dark' : 'light', accent_watch_supported: false }),
    [S99 + 'title.Service.Set']: (args) => {
      document.title = String(args[0] ?? '')
      return null
    },
    [S99 + 'title.Service.Title']: () => document.title,
    [S99 + 'browser.Service.OpenURL']: (args) => {
      window.open(String(args[0] ?? ''), '_blank', 'noopener')
      return null
    },
    // The desktop log goes to a file; here the console is the log.
    /* eslint-disable no-console */
    [S99 + 'logger.Service.LogDebug']: args => void console.debug(...args),
    [S99 + 'logger.Service.LogInfo']: args => void console.info(...args),
    [S99 + 'logger.Service.LogWarn']: args => void console.warn(...args),
    [S99 + 'logger.Service.LogError']: args => void console.error(...args),
    /* eslint-enable no-console */
  }
  return new Map(Object.entries(table).map(([fqn, fn]) => [bindingID(fqn), fn]))
}
