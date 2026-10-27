// app/web/useWebEngine.ts
// The web edition's engine and microphone: one of each per page, shared by the shell, the gate and the diagnostics.
import { ref, shallowRef } from 'vue'
import { Events } from '@wailsio/runtime'
import * as ConfigService from '~~bindings/smegg.me/smeggtuner/backend/services/config/service.js'
import { bootEngine, type WebEngine } from './engine'
import { localServices, type MicControl } from './local'
import { MicError, appliedProcessing, requestMic, type MicReason } from './mic'
import { PERF_EVENT, type PerfDTO } from './protocol'
import { loadConfig, storedDeviceId } from './storage'

const status = ref<'loading' | 'ready' | 'failed'>('loading')
const progress = ref(0)
const failure = ref('')
const micError = ref<MicReason | null>(null)
const micOpen = ref(false)
const processing = ref<string[]>([])
const track = shallowRef<MediaTrackSettings | null>(null)
const perf = ref<PerfDTO | null>(null)

let engine: WebEngine | null = null
let stream: MediaStream | null = null

const mic: MicControl & { deviceId: string, label: string } = {
  deviceId: '',
  label: '',
  async select(deviceId: string) {
    mic.deviceId = deviceId
    mic.label = ''
    if (stream) await openMic()
  },
}

function closeMic() {
  engine?.detach()
  stream?.getTracks().forEach(t => t.stop())
  stream = null
  micOpen.value = false
}

async function openMic(): Promise<boolean> {
  if (!engine) return false
  closeMic()
  try {
    stream = await requestMic(mic.deviceId)
  }
  catch (e) {
    // A saved microphone that is gone falls back to the default, as the desktop app does.
    if (!(e instanceof MicError && e.reason === 'none' && mic.deviceId)) {
      micError.value = e instanceof MicError ? e.reason : 'failed'
      return false
    }
    mic.deviceId = ''
    mic.label = ''
    return openMic()
  }
  micError.value = null
  const t = stream.getAudioTracks()[0]
  mic.label = t?.label ?? ''
  track.value = t?.getSettings() ?? null
  processing.value = appliedProcessing(stream)
  engine.attach(stream)
  micOpen.value = true
  return true
}

async function boot() {
  if (engine || status.value === 'failed') return
  mic.deviceId = storedDeviceId(loadConfig())
  try {
    engine = await bootEngine((loaded, total) => {
      progress.value = total ? Math.min(1, loaded / total) : 0
    }, localServices(mic, id => ConfigService.PatchConfig({ audio: { device_id: id } })))
    resumeOnGesture()
    Events.On(PERF_EVENT, (e: { data: PerfDTO }) => {
      perf.value = e.data
    })
    status.value = 'ready'
  }
  catch (e) {
    failure.value = String(e)
    status.value = 'failed'
  }
}

// Audio may only resume from a user gesture, and iOS wants it inside the handler; any tap or key does, the run key included.
function resumeOnGesture() {
  const resume = () => {
    if (engine?.context.state === 'suspended') void engine.context.resume()
  }
  for (const type of ['pointerdown', 'keydown'] as const) window.addEventListener(type, resume, { capture: true })
}

export function useWebEngine() {
  return {
    status,
    progress,
    failure,
    micError,
    micOpen,
    processing,
    track,
    perf,
    boot,
    openMic,
    closeMic,
    setDevice: (id: string) => { mic.deviceId = id },
    context: () => engine?.context ?? null,
  }
}
