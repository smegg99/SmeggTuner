import { ref, shallowRef } from 'vue'
import * as AudioService from '~~bindings/smegg.me/smeggtuner/services/audio/service.js'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/services/tuner/service.js'
import { SourceKind } from '~~bindings/smegg.me/smeggtuner/services/audio/models.js'
import type { DeviceDTO, SourceDTO } from '~~bindings/smegg.me/smeggtuner/services/audio/models.js'
import { toErrorKey, useTuner } from '~/composables/useTuner'
import { useTransport } from '~/composables/useTransport'

export interface AudioDeviceView { id: string, name: string, default: boolean }

export interface AudioSourceView {
  kind: 'mic' | 'file'
  deviceId: string
  path: string
  loop: boolean
  name: string
}

const NO_SOURCE: AudioSourceView = { kind: 'mic', deviceId: '', path: '', loop: false, name: '' }

const devices = shallowRef<AudioDeviceView[]>([])
const current = ref<AudioSourceView>({ ...NO_SOURCE })
const loading = ref(false)
const error = ref('')

let started = false
// Device enumeration is slow enough to overlap with a second refresh; only the
// newest one may write.
let refreshSeq = 0

type Logger = ReturnType<typeof useLogger>

function toSourceView(dto: SourceDTO): AudioSourceView {
  return {
    kind: dto.kind === SourceKind.SourceFile ? 'file' : 'mic',
    deviceId: dto.deviceId,
    path: dto.path,
    loop: dto.loop,
    name: dto.name,
  }
}

async function refresh(log: Logger) {
  const seq = ++refreshSeq
  loading.value = true

  try {
    const [list, source]: [DeviceDTO[], SourceDTO] = await Promise.all([
      AudioService.ListDevices(),
      AudioService.Current(),
    ])
    if (seq !== refreshSeq) return

    devices.value = list
    current.value = toSourceView(source)
    error.value = ''

    // A recording from a prior session returns with the config but nothing re-selects it; without this refresh the file view starts empty.
    await useTransport().refresh()
  }
  catch (err) {
    if (seq !== refreshSeq) return

    devices.value = []
    error.value = toErrorKey(err)
    log.error('audio: failed to list devices', { error: String(err) })
  }
  finally {
    if (seq === refreshSeq) loading.value = false
  }
}

// A Source is single-Start, so switching input restarts a running engine rather than reusing it.
// The service validates and rejects invalid selections, so `current` only reflects backend state.
async function select(call: () => Promise<void>, log: Logger) {
  loading.value = true

  try {
    await call()
    error.value = ''
    current.value = toSourceView(await AudioService.Current())

    // Clear every reading: it was all measured from the replaced source and would otherwise read as the new one.
    useTuner().clearReading()

    // Single choke point for source changes; ask Go whether a transport exists rather than inferring it from the path.
    await useTransport().refresh()

    if (await TunerService.IsRunning()) {
      await TunerService.Restart()
    }
  }
  catch (err) {
    error.value = toErrorKey(err)
    log.error('audio: failed to select source', { error: String(err) })
  }
  finally {
    loading.value = false
  }
}

// Title/filter are passed in because this module-scope code has no useI18n, and Go can't localize them (see services/audio.OpenFileDialog).
interface DialogWords {
  title: string
  filter: string
}

async function pickFile(log: Logger, words: DialogWords) {
  let path = ''

  try {
    path = await AudioService.OpenFileDialog(words.title, words.filter)
  }
  catch (err) {
    error.value = toErrorKey(err)
    log.error('audio: file dialog failed', { error: String(err) })
    return
  }

  if (!path) return

  await select(() => AudioService.SelectFile(path, current.value.loop), log)
}

export function useAudioDevices() {
  const log = useLogger()
  const { t } = useI18n()

  if (!started) {
    started = true
    void refresh(log)
  }

  return {
    devices,
    current,
    loading,
    error,
    refresh: () => refresh(log),
    selectMic: (deviceId: string) => select(() => AudioService.SelectMic(deviceId), log),
    selectFile: (path: string, loop: boolean) => select(() => AudioService.SelectFile(path, loop), log),
    pickFile: () => pickFile(log, { title: t('file.openTitle'), filter: t('file.openFilter') }),
  }
}
